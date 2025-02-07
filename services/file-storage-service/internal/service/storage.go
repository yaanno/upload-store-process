package service

import (
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
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

type FileStorageService interface {
	DeleteFile(context.Context, *storagev1.DeleteFileRequest) (*storagev1.DeleteFileResponse, error)
	GetFileMetadata(context.Context, *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error)
	ListFiles(context.Context, *storagev1.ListFilesRequest) (*storagev1.ListFilesResponse, error)
	PrepareUpload(context.Context, *storagev1.PrepareUploadRequest) (*storagev1.PrepareUploadResponse, error)
}

// FileStorageService implements the gRPC service
type fileStorageService struct {
	storagev1.UnimplementedFileStorageServiceServer
	repo            repository.FileMetadataRepository
	logger          logger.Logger
	storageProvider storage.FileStorageProvider
}

// CompressFile implements v1.FileStorageServiceServer.
// func (s *fileStorageService) CompressFile(context.Context, *storagev1.CompressFileRequest) (*storagev1.CompressFileResponse, error) {
// 	return nil, status.Error(codes.Unimplemented, "method CompressFile not implemented")
// }

// DeleteFile implements v1.FileStorageServiceServer.
func (s *fileStorageService) DeleteFile(context.Context, *storagev1.DeleteFileRequest) (*storagev1.DeleteFileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteFile not implemented")
}

// GetFileMetadata implements v1.FileStorageServiceServer.
func (s *fileStorageService) GetFileMetadata(context.Context, *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetFileMetadata not implemented")
}

// StoreFileMetadata implements v1.FileStorageServiceServer.
// func (s *fileStorageService) StoreFileMetadata(context.Context, *storagev1.StoreFileMetadataRequest) (*storagev1.StoreFileMetadataResponse, error) {
// 	return nil, status.Error(codes.Unimplemented, "method StoreFileMetadata not implemented")
// }

// NewFileStorageService creates a new instance of FileStorageService
func NewFileStorageService(
	repo repository.FileMetadataRepository,
	logger logger.Logger,
	storageProvider storage.FileStorageProvider,
) *fileStorageService {
	return &fileStorageService{
		repo:            repo,
		logger:          logger,
		storageProvider: storageProvider,
	}
}

// PrepareUpload prepares a file upload by storing initial metadata
func (s *fileStorageService) PrepareUpload(
	ctx context.Context,
	req *storagev1.PrepareUploadRequest,
) (*storagev1.PrepareUploadResponse, error) {
	// Validate input
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "upload request cannot be nil")
	}

	// TODO: implement JWT token validation as a middleware
	// if req.JwtToken == "" {
	// 	return nil, status.Errorf(codes.InvalidArgument, "JWT token is required")
	// }

	if req.GlobalUploadId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "global upload ID is required")
	}

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
		UserId:           extractUserID(ctx), // Implement user context extraction
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

// CompleteUpload finalizes the file upload process
// func (s *fileStorageService) CompleteUpload(
// 	ctx context.Context,
// 	req *storagev1.CompleteUploadRequest,
// ) (*storagev1.CompleteUploadResponse, error) {
// 	// Validate input
// 	if req == nil {
// 		return nil, status.Errorf(codes.InvalidArgument, "complete upload request cannot be nil")
// 	}

// 	// Validate upload ID
// 	if req.UploadId == "" {
// 		return nil, status.Errorf(codes.InvalidArgument, "upload ID is required")
// 	}

// 	// Validate file metadata
// 	if req.FileMetadata == nil {
// 		return nil, status.Errorf(codes.InvalidArgument, "file metadata is required")
// 	}

// 	// Prepare metadata for storage
// 	metadata := &models.FileMetadataRecord{
// 		ID:               req.UploadId, // Use the provided upload ID
// 		ProcessingStatus: "COMPLETED",
// 		Metadata:         req.FileMetadata,
// 	}

// 	// Store metadata
// 	err := s.repo.CreateFileMetadata(ctx, metadata)
// 	if err != nil {
// 		s.logger.Error().Err(err).Msg("Failed to store file metadata")
// 		return nil, status.Errorf(codes.Internal, "failed to store file metadata: %v", err)
// 	}

// 	return &storagev1.CompleteUploadResponse{
// 		ProcessedFileId:   req.UploadId, // Return the same upload ID
// 		ProcessingStarted: true,
// 		BaseResponse: &sharedv1.Response{
// 			Message: "Upload completed successfully",
// 		},
// 	}, nil
// }

// ListFiles retrieves a list of files for a user
func (s *fileStorageService) ListFiles(
	ctx context.Context,
	req *storagev1.ListFilesRequest,
) (*storagev1.ListFilesResponse, error) {
	// Validate input
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "list files request cannot be nil")
	}

	userID := extractUserID(ctx)
	if userID == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	// Default page size and page number if not provided
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 1 // Default to 10 if not specified
	}
	pageNum := req.Page
	if pageNum < 1 {
		pageNum = 1
	}

	// Prepare list options
	listOpts := &repository.FileMetadataListOptions{
		UserID: userID,
		Limit:  int(pageSize),
		Offset: int((pageNum - 1) * pageSize),
	}

	// Retrieve file metadata
	fileMetadataList, err := s.repo.ListFileMetadata(ctx, listOpts)
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
		totalPages = int32(math.Ceil(float64(len(fileMetadataList)) / float64(pageSize)))
	}

	return &storagev1.ListFilesResponse{
		Files:      files,
		TotalFiles: int32(len(fileMetadataList)),
		TotalPages: totalPages,
	}, nil
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

func extractUserID(ctx context.Context) string {
	// Implement user ID extraction from context
	// This is a placeholder - you'd typically use authentication middleware
	return "test-user"
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
func validateUploadToken(token string, fileID string) bool {
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

var _ FileStorageService = &fileStorageService{}
