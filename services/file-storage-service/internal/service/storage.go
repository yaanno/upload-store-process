package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"path/filepath"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/repository"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage"
	"github.com/yaanno/upload-store-process/services/shared/pkg/auth"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

type FileStorageService interface {
	DeleteFile(context.Context, *storagev1.DeleteFileRequest) (*storagev1.DeleteFileResponse, error)
	GetFileMetadata(context.Context, *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error)
	ListFiles(context.Context, *storagev1.ListFilesRequest) (*storagev1.ListFilesResponse, error)
	PrepareUpload(context.Context, *storagev1.PrepareUploadRequest) (*storagev1.PrepareUploadResponse, error)
	UploadFile(context.Context, *storagev1.UploadFileRequest) (*storagev1.UploadFileResponse, error)
}

// FileStorageService implements the gRPC service
type FileStorageServiceImpl struct {
	storagev1.UnimplementedFileStorageServiceServer
	repo            repository.FileMetadataRepository
	logger          logger.Logger
	storageProvider storage.FileStorageProvider
	tokenValidator  auth.TokenValidator
}

// NewFileStorageService creates a new instance of FileStorageService
func NewFileStorageService(
	repo repository.FileMetadataRepository,
	logger logger.Logger,
	storageProvider storage.FileStorageProvider,
	tokenValidator auth.TokenValidator,
) *FileStorageServiceImpl {
	return &FileStorageServiceImpl{
		repo:            repo,
		logger:          logger,
		storageProvider: storageProvider,
		tokenValidator:  tokenValidator,
	}
}

// PrepareUpload prepares a file upload by storing initial metadata
func (s *FileStorageServiceImpl) PrepareUpload(
	ctx context.Context,
	req *storagev1.PrepareUploadRequest,
) (*storagev1.PrepareUploadResponse, error) {

	if err := s.ValidatePrepareUploadRequest(ctx, req); err != nil {
		return nil, err
	}

	// Generate secure file ID
	fileID, err := generateSecureFileID()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate file ID")
		return nil, status.Errorf(codes.Internal, "failed to generate file ID")
	}

	// Generate upload token
	uploadToken, err := generateSecureUploadToken(fileID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate upload token")
		return nil, status.Errorf(codes.Internal, "failed to generate upload token")
	}

	// Generate storage path using the storage provider
	storagePath := s.storageProvider.GenerateStoragePath(fileID, req.Filename)

	// Prepare initial metadata
	initialMetadata := &sharedv1.FileMetadata{
		FileId:           fileID,
		OriginalFilename: req.Filename,
		FileSizeBytes:    req.FileSizeBytes,
		FileType:         determineFileType(req.Filename),
		UploadTimestamp:  time.Now().Unix(),
		UserId:           req.UserId,
	}

	// Create initial metadata record
	metadataRecord := &models.FileMetadataRecord{
		ID:               fileID,
		Metadata:         initialMetadata,
		StoragePath:      storagePath,
		ProcessingStatus: "PENDING",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	// Store initial metadata
	if err := s.repo.CreateFileMetadata(ctx, metadataRecord); err != nil {
		s.logger.Error().Err(err).Msg("Failed to create initial file metadata")
		return nil, status.Errorf(codes.Internal, "failed to create file metadata")
	}

	return &storagev1.PrepareUploadResponse{
		StorageUploadToken: uploadToken,
		StoragePath:        storagePath,
		BaseResponse: &sharedv1.Response{
			Message: "Upload prepared successfully",
		},
	}, nil
}

// ListFiles retrieves a list of files for a user
func (s *FileStorageServiceImpl) ListFiles(ctx context.Context, req *storagev1.ListFilesRequest) (*storagev1.ListFilesResponse, error) {
	// Validate input
	if err := s.ValidateListFilesRequest(ctx, req); err != nil {
		return nil, err
	}

	// Calculate pagination
	pageSize := int32(req.PageSize)
	pageNum := int32(req.Page)

	// Prepare list options
	listOpts := &models.FileMetadataListOptions{
		UserID: req.UserId,
		Limit:  int(pageSize),
		Offset: int((pageNum - 1) * pageSize),
	}

	// Retrieve file metadata
	fileMetadataList, totalCount, err := s.repo.ListFiles(ctx, listOpts)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list file metadata")
		return nil, status.Errorf(codes.Internal, "failed to list files")
	}

	// Convert to gRPC response
	var files []*sharedv1.FileMetadata
	for _, metadata := range fileMetadataList {
		files = append(files, &sharedv1.FileMetadata{
			FileId:           metadata.ID,
			OriginalFilename: metadata.Metadata.OriginalFilename,
			FileSizeBytes:    metadata.Metadata.FileSizeBytes,
			UploadTimestamp:  metadata.Metadata.UploadTimestamp,
		})
	}

	// Calculate total pages
	var totalPages int32
	if pageSize > 0 {
		totalPages = int32(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	return &storagev1.ListFilesResponse{
		Files:      files,
		TotalFiles: int32(totalCount),
		TotalPages: totalPages,
	}, nil
}

// UploadFile handles the actual file upload process
func (s *FileStorageServiceImpl) UploadFile(
	ctx context.Context,
	req *storagev1.UploadFileRequest,
) (*storagev1.UploadFileResponse, error) {
	// Validate input
	if err := s.ValidateUploadFileRequest(ctx, req); err != nil {
		return nil, err
	}

	// Retrieve file metadata to confirm upload context
	metadata, err := s.repo.RetrieveFileMetadataByID(ctx, req.FileId)
	if err != nil {
		s.logger.Error().
			Str("method", "UploadFile").
			Err(err).
			Str("fileId", req.FileId).
			Msg("failed to retrieve file metadata")
		return nil, status.Errorf(codes.NotFound, "file metadata not found")
	}

	// Validate that file is in a valid state for upload
	if metadata.ProcessingStatus != "PENDING" {
		s.logger.Error().
			Str("method", "UploadFile").
			Str("fileId", req.FileId).
			Str("processingStatus", metadata.ProcessingStatus).
			Msg("invalid file upload state")
		return nil, status.Errorf(codes.FailedPrecondition, "invalid file upload state")
	}

	// Update metadata to UPLOADING status
	metadata.ProcessingStatus = "UPLOADING"
	if err := s.repo.UpdateFileMetadata(ctx, metadata); err != nil {
		s.logger.Error().
			Str("method", "UploadFile").
			Err(err).
			Str("fileId", req.FileId).
			Msg("failed to update file metadata")
		return nil, status.Errorf(codes.Internal, "failed to update file metadata")
	}

	// Store the file using the storage provider
	storagePath, err := s.storageProvider.StoreFile(
		ctx,
		req.FileId,
		metadata.Metadata.OriginalFilename,
		bytes.NewReader(req.FileContent),
	)
	if err != nil {
		// Rollback metadata status
		metadata.ProcessingStatus = "PENDING"
		if err := s.repo.UpdateFileMetadata(ctx, metadata); err != nil {
			s.logger.Error().
				Str("method", "UploadFile").
				Err(err).
				Str("fileId", req.FileId).
				Msg("failed to rollback file metadata")
		}
		return nil, status.Errorf(codes.Internal, "failed to store file: %v", err)
	}

	// Update metadata to COMPLETED status
	metadata.ProcessingStatus = "COMPLETED"
	metadata.StoragePath = storagePath
	metadata.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateFileMetadata(ctx, metadata); err != nil {
		s.logger.Error().
			Str("method", "UploadFile").
			Str("fileID", req.FileId).
			Err(err).
			Msg("Failed to update file metadata after upload")
		// Non-critical error, file is already stored
	}

	return &storagev1.UploadFileResponse{
		BaseResponse: &sharedv1.Response{
			Message: "File uploaded successfully",
		},
		StoragePath: storagePath,
	}, nil
}

// DeleteFile implements v1.FileStorageServiceServer.
func (s *FileStorageServiceImpl) DeleteFile(ctx context.Context, req *storagev1.DeleteFileRequest) (*storagev1.DeleteFileResponse, error) {

	if err := s.ValidateDeleteFileRequest(ctx, req); err != nil {
		return nil, err
	}

	//  delete file from database
	if err := s.repo.SoftDeleteFile(ctx, req.FileId, req.UserId); err != nil {
		s.logger.Error().
			Str("method", "DeleteFile").
			Err(err).
			Str("fileId", req.FileId).
			Msg("failed to soft delete file metadata")
		return nil, status.Errorf(codes.Internal, "failed to soft delete file metadata: %v", err)
	}

	return &storagev1.DeleteFileResponse{
		FileDeleted: true,
		DeletedAt:   timestamppb.Now(),
		BaseResponse: &sharedv1.Response{
			Message: "File deleted successfully",
		},
	}, nil
}

// GetFileMetadata implements v1.FileStorageServiceServer.
func (s *FileStorageServiceImpl) GetFileMetadata(ctx context.Context, req *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error) {
	if err := s.ValidateGetFileMetadataRequest(ctx, req); err != nil {
		return nil, err
	}

	// Retrieve file metadata
	metadata, err := s.repo.RetrieveFileMetadataByID(ctx, req.FileId)
	if err != nil {
		s.logger.Error().
			Str("method", "GetFileMetadata").
			Err(err).
			Str("fileId", req.FileId).
			Msg("failed to retrieve file metadata")
		return nil, status.Errorf(codes.NotFound, "file metadata not found")
	}

	// Transform metadata to GetFileMetadataResponse
	fileMetadata := &sharedv1.FileMetadata{
		FileId:           metadata.ID,
		OriginalFilename: metadata.Metadata.OriginalFilename,
		FileSizeBytes:    metadata.Metadata.FileSizeBytes,
		FileType:         metadata.Metadata.FileType,
		UploadTimestamp:  metadata.Metadata.UploadTimestamp,
		UserId:           metadata.Metadata.UserId,
		StoragePath:      metadata.StoragePath,
	}

	// Return response
	return &storagev1.GetFileMetadataResponse{
		BaseResponse: &sharedv1.Response{
			Message: "File metadata retrieved successfully",
		},
		Metadata: fileMetadata,
	}, nil
}

// Request validators

// ValidateUploadRequest validates the upload request
func (s *FileStorageServiceImpl) ValidatePrepareUploadRequest(ctx context.Context, req *storagev1.PrepareUploadRequest) error {
	if req == nil {
		return status.Errorf(codes.InvalidArgument, "request cannot be nil")
	}
	// Validate user ID
	if req.UserId == "" {
		s.logger.Error().
			Str("method", "ValidateUploadRequest").
			Str("user_id", req.UserId).
			Msg("user ID is empty")
		return status.Errorf(codes.InvalidArgument, "user ID is empty")
	}

	// Validate file name
	if req.Filename == "" {
		s.logger.Error().
			Str("method", "ValidateUploadRequest").
			Str("filename", req.Filename).
			Msg("filename is empty")
		return status.Errorf(codes.InvalidArgument, "filename is empty")
	}
	// Validate file size
	if req.FileSizeBytes <= 0 {
		s.logger.Error().
			Str("method", "ValidateUploadRequest").
			Int64("filesize", req.FileSizeBytes).
			Msg("filesize is empty")
		return status.Errorf(codes.InvalidArgument, "filesize is empty")
	}
	// Validate file size
	if req.FileSizeBytes > 500*1024*1024 { // 500 MB
		s.logger.Error().
			Str("method", "PrepareUpload").
			Str("file_size_bytes", strconv.FormatInt(req.FileSizeBytes, 10)).
			Msg("file too large")
		return status.Errorf(codes.InvalidArgument, "file too large")
	}
	// Validate file type
	if !isAllowedFileType(req.Filename) {
		s.logger.Error().
			Str("method", "PrepareUpload").
			Str("filename", req.Filename).
			Msg("unsupported file type")
		return status.Errorf(codes.InvalidArgument, "unsupported file type")
	}
	return nil
}

// ValidateUploadRequest validates the upload request
func (s *FileStorageServiceImpl) ValidateUploadFileRequest(ctx context.Context, req *storagev1.UploadFileRequest) error {
	if req == nil {
		s.logger.Error().
			Str("method", "UploadFile").
			Msg("upload request cannot be nil")
		return status.Errorf(codes.InvalidArgument, "upload request cannot be nil")
	}
	// TODO: Validate JWT token

	// Validate upload token
	if !s.IsUploadTokenValid(req.StorageUploadToken, req.FileId) {
		s.logger.Error().
			Str("method", "UploadFile").
			Msg("invalid upload token")
		return status.Errorf(codes.PermissionDenied, "invalid upload token")
	}
	return nil
}

// ValidateDeleteRequest validates the delete request
func (s *FileStorageServiceImpl) ValidateDeleteFileRequest(ctx context.Context, req *storagev1.DeleteFileRequest) error {
	if req == nil {
		s.logger.Error().
			Str("method", "DeleteFile").
			Msg("delete request cannot be nil")
		return status.Errorf(codes.InvalidArgument, "delete request cannot be nil")
	}
	// Validate user ID
	if req.UserId == "" {
		s.logger.Error().
			Str("method", "DeleteFile").
			Str("user_id", req.UserId).
			Msg("user ID is empty")
		return status.Errorf(codes.InvalidArgument, "user ID is required")
	}

	// Validate file ID
	if req.FileId == "" {
		s.logger.Error().
			Str("method", "DeleteFile").
			Str("fileId", req.FileId).
			Msg("file ID cannot be empty")
		return status.Errorf(codes.InvalidArgument, "file ID cannot be empty")
	}

	opts := &models.FileMetadataListOptions{
		UserID: req.UserId,
		FileID: req.FileId,
	}

	isOwner, err := s.repo.IsFileOwnedByUser(ctx, opts)
	if err != nil {
		s.logger.Error().
			Str("method", "DeleteFile").
			Err(err).
			Str("fileId", req.FileId).
			Msg("failed to check file ownership")
		return status.Errorf(codes.Internal, "failed to check file ownership: %v", err)
	}

	if !isOwner && !req.ForceDelete {
		s.logger.Warn().
			Str("method", "DeleteFile").
			Str("fileId", req.FileId).
			Str("userId", req.UserId).
			Msg("user does not own file")
		return status.Errorf(codes.PermissionDenied, "user does not own file")
	}

	return nil
}

// ValidateListFilesRequest validates the list files request
func (s *FileStorageServiceImpl) ValidateListFilesRequest(ctx context.Context, req *storagev1.ListFilesRequest) error {
	if req == nil {
		s.logger.Error().
			Str("method", "ListFiles").
			Msg("list files request cannot be nil")
		return status.Errorf(codes.InvalidArgument, "list files request cannot be nil")
	}

	if req.UserId == "" {
		s.logger.Error().
			Str("method", "ListFiles").
			Str("user_id", req.UserId).
			Msg("user ID is empty")
		return status.Errorf(codes.InvalidArgument, "user ID is required")
	}

	if req.Page < 1 {
		s.logger.Error().
			Str("method", "ListFiles").
			Int32("page", req.Page).
			Msg("page must be at least 1")
		return status.Errorf(codes.InvalidArgument, "page must be at least 1")
	}

	if req.PageSize <= 0 {
		s.logger.Error().
			Str("method", "ListFiles").
			Int32("page_size", req.PageSize).
			Msg("page size must be at least 1")
		return status.Errorf(codes.InvalidArgument, "page size must be at least 1")
	}

	return nil
}

// ValidateGetFileMetadataRequest validates the get file metadata request
func (s *FileStorageServiceImpl) ValidateGetFileMetadataRequest(ctx context.Context, req *storagev1.GetFileMetadataRequest) error {
	if req == nil {
		s.logger.Error().
			Str("method", "GetFileMetadata").
			Msg("get file metadata request cannot be nil")
		return status.Errorf(codes.InvalidArgument, "get file metadata request cannot be nil")
	}

	if req.UserId == "" {
		s.logger.Error().
			Str("method", "GetFileMetadata").
			Str("user_id", req.UserId).
			Msg("user ID is empty")
		return status.Errorf(codes.InvalidArgument, "user ID is required")
	}

	if req.FileId == "" {
		s.logger.Error().
			Str("method", "GetFileMetadata").
			Str("file_id", req.FileId).
			Msg("file ID cannot be empty")
		return status.Errorf(codes.InvalidArgument, "file ID cannot be empty")
	}

	opts := &models.FileMetadataListOptions{
		UserID: req.UserId,
		FileID: req.FileId,
	}

	isOwner, err := s.repo.IsFileOwnedByUser(ctx, opts)
	if err != nil {
		s.logger.Error().
			Str("method", "GetFileMetadata").
			Err(err).
			Str("file_id", req.FileId).
			Msg("failed to check file ownership")
		return status.Errorf(codes.Internal, "failed to check file ownership: %v", err)
	}

	if !isOwner {
		s.logger.Warn().
			Str("method", "GetFileMetadata").
			Str("file_id", req.FileId).
			Str("user_id", req.UserId).
			Msg("user does not own file")
		return status.Errorf(codes.PermissionDenied, "user does not own file")
	}

	return nil
}

// Utility functions

// generateSecureFileID creates a cryptographically secure, unique file identifier
func generateSecureFileID() (string, error) {
	// Generate 32 bytes of cryptographically secure random data
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random bytes: %w", err)
	}

	// Create a hash to add an extra layer of unpredictability
	hash := sha256.Sum256(append(randomBytes, []byte(time.Now().String())...))

	// Use URL-safe base64 encoding to ensure safe use in URLs and file systems
	return base64.URLEncoding.EncodeToString(hash[:]), nil
}

// generateSecureUploadToken creates a time-limited, secure upload token
func generateSecureUploadToken(fileID string) (string, error) {
	// Generate 64 bytes of cryptographically secure random data
	tokenBytes := make([]byte, 64)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}

	// Create a hash that includes file ID, random bytes, and current timestamp
	tokenData := append(tokenBytes, []byte(fileID)...)
	tokenData = append(tokenData, []byte(time.Now().String())...)

	hash := sha256.Sum256(tokenData)

	// Combine timestamp and base64 encoded hash for additional security
	token := fmt.Sprintf("%d_%s",
		time.Now().Add(time.Hour).Unix(), // Token expires in 1 hour
		base64.URLEncoding.EncodeToString(hash[:]),
	)

	return token, nil
}

func determineFileType(filename string) string {
	// Implement basic file type detection
	// This is a very simple implementation, consider using a more robust method
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

func isAllowedFileType(filename string) bool {
	allowedExtensions := []string{".csv", ".json", ".txt"}
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

// Optional: Token validation function
func (s *FileStorageServiceImpl) IsUploadTokenValid(token string, fileID string) bool {
	if token == "" {
		return false
	}

	if fileID == "" {
		return false
	}

	parts := strings.Split(token, "_")
	if len(parts) != 2 {
		return false
	}

	// Check token expiration
	expirationTime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix() > expirationTime {
		return false // Token expired
	}

	// Optionally, you could add additional validation logic here
	return true
}

var _ FileStorageService = &FileStorageServiceImpl{}
