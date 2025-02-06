package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"path/filepath"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/repository"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

// FileStorageService implements the gRPC service
type FileStorageService struct {
	storagev1.UnimplementedFileStorageServiceServer
	repo   repository.FileMetadataRepository
	logger *logger.Logger
}

// NewFileStorageService creates a new instance of FileStorageService
func NewFileStorageService(
	repo repository.FileMetadataRepository,
	logger *logger.Logger,
) *FileStorageService {
	return &FileStorageService{
		repo:   repo,
		logger: logger,
	}
}

// PrepareUpload prepares a file upload by storing initial metadata
func (s *FileStorageService) PrepareUpload(
	ctx context.Context,
	req *storagev1.PrepareUploadRequest,
) (*storagev1.PrepareUploadResponse, error) {
	// Validate input
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "upload request cannot be nil")
	}

	if req.Filename == "" {
		return nil, status.Errorf(codes.InvalidArgument, "filename is required")
	}

	if req.FileSizeBytes <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid file size")
	}

	// Generate a unique file ID (you might want to use a more robust ID generation method)
	fileID := generateFileID()

	// Prepare initial metadata
	initialMetadata := &sharedv1.FileMetadata{
		FileId:           fileID,
		OriginalFilename: req.Filename,
		FileSizeBytes:    req.FileSizeBytes,
		FileType:         determineFileType(req.Filename),
		UploadTimestamp:  time.Now().Unix(),
		UserId:           extractUserID(ctx), // Implement user context extraction
	}

	// Convert to internal model
	storageModel := &models.FileMetadataRecord{
		ID:               fileID,
		Metadata:         initialMetadata,
		StoragePath:      generateStoragePath(fileID, req.Filename),
		ProcessingStatus: "PENDING",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	// Store initial metadata
	if err := s.repo.CreateFileMetadata(ctx, storageModel); err != nil {
		s.logger.Error().Err(err).Msg("Failed to store file metadata")
		return nil, status.Errorf(codes.Internal, "failed to prepare upload")
	}

	// Generate a presigned URL or upload token (implementation depends on your storage strategy)
	presignedURL, err := generatePresignedURL(fileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate upload URL")
	}

	return &storagev1.PrepareUploadResponse{
		UploadToken: presignedURL,
		StoragePath: generateStoragePath(fileID, req.Filename),
	}, nil
}

// CompleteUpload finalizes the file upload process
func (s *FileStorageService) CompleteUpload(
	ctx context.Context,
	req *storagev1.CompleteUploadRequest,
) (*storagev1.CompleteUploadResponse, error) {
	// Validate input
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "complete upload request cannot be nil")
	}

	if req.UploadId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "upload ID is required")
	}

	// Retrieve existing metadata
	existingMetadata, err := s.repo.RetrieveFileMetadataByID(ctx, req.UploadId)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to retrieve file metadata")
		return nil, status.Errorf(codes.NotFound, "upload not found")
	}

	// Update metadata with final details
	existingMetadata.ProcessingStatus = "COMPLETED"
	existingMetadata.UpdatedAt = time.Now().UTC()

	// Update metadata in repository
	if err := s.repo.CreateFileMetadata(ctx, existingMetadata); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update file metadata")
		return nil, status.Errorf(codes.Internal, "failed to complete upload")
	}

	return &storagev1.CompleteUploadResponse{
		ProcessedFileId:   existingMetadata.ID,
		ProcessingStarted: true,
	}, nil
}

// ListFiles retrieves a list of files for a user
func (s *FileStorageService) ListFiles(
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

// Utility functions (these would typically be in separate utility packages)
func generateFileID() string {
	// Implement a robust ID generation method
	return fmt.Sprintf("file_%d", time.Now().UnixNano())
}

func generateStoragePath(fileID, filename string) string {
	// Implement storage path generation logic
	return fmt.Sprintf("/uploads/%s/%s", fileID, filename)
}

func generatePresignedURL(fileID string) (string, error) {
	// Implement presigned URL generation
	return fmt.Sprintf("https://example.com/upload/%s", fileID), nil
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
