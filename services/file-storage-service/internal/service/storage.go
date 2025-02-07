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
	// Validate input
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "upload request cannot be nil")
	}

	// Validate JWT token
	claims, err := s.tokenValidator.ValidateToken(req.JwtToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid JWT token: %v", err)
	}

	// Validate user ID
	if claims.UserID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user ID is required")
	}

	// Validate global upload ID
	if req.GlobalUploadId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "global upload ID is required")
	}

	// Validate filename
	if req.Filename == "" {
		return nil, status.Errorf(codes.InvalidArgument, "filename is required")
	}

	if req.FileSizeBytes <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid file size")
	}

	// Validate file type
	if !isAllowedFileType(req.Filename) {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported file type")
	}

	// Validate file size
	if req.FileSizeBytes > 500*1024*1024 { // 500 MB
		return nil, status.Errorf(codes.InvalidArgument, "file too large")
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
		UserId:           claims.UserID,
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
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "list files request cannot be nil")
	}

	// Validate JWT token
	claims, err := s.tokenValidator.ValidateToken(req.JwtToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid JWT token: %v", err)
	}

	// Validate pagination parameters
	pageNum := req.Page
	if pageNum < 1 {
		pageNum = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	// Prepare list options
	listOpts := &repository.FileMetadataListOptions{
		UserID: claims.UserID,
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
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "upload request cannot be nil")
	}

	// Validate upload token
	if !s.IsUploadTokenValid(req.StorageUploadToken, req.FileId) {
		return nil, status.Errorf(codes.PermissionDenied, "invalid upload token")
	}

	// Retrieve file metadata to confirm upload context
	metadata, err := s.repo.RetrieveFileMetadataByID(ctx, req.FileId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "file metadata not found")
	}

	// Validate that file is in a valid state for upload
	if metadata.ProcessingStatus != "PENDING" {
		return nil, status.Errorf(codes.FailedPrecondition, "invalid file upload state")
	}

	// Update metadata to UPLOADING status
	metadata.ProcessingStatus = "UPLOADING"
	if err := s.repo.UpdateFileMetadata(ctx, metadata); err != nil {
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
		_ = s.repo.UpdateFileMetadata(ctx, metadata)
		return nil, status.Errorf(codes.Internal, "failed to store file: %v", err)
	}

	// Update metadata to COMPLETED status
	metadata.ProcessingStatus = "COMPLETED"
	metadata.StoragePath = storagePath
	metadata.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateFileMetadata(ctx, metadata); err != nil {
		s.logger.Error().
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
func (s *FileStorageServiceImpl) DeleteFile(context.Context, *storagev1.DeleteFileRequest) (*storagev1.DeleteFileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteFile not implemented")
}

// GetFileMetadata implements v1.FileStorageServiceServer.
func (s *FileStorageServiceImpl) GetFileMetadata(context.Context, *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetFileMetadata not implemented")
}

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
