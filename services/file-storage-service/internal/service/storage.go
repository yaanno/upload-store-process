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

	// Validate file size
	if req.FileSizeBytes > 500*1024*1024 { // 500 MB
		return nil, status.Errorf(codes.InvalidArgument, "file too large")
	}

	// Validate file type
	if !isAllowedFileType(req.Filename) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid file type")
	}

	// Generate a unique file ID (you might want to use a more robust ID generation method)
	fileID, err := generateSecureFileID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate file ID: %v", err)
	}

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
	uploadToken, err := generateSecureUploadToken(fileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate upload URL")
	}

	// TODO: add response fields
	return &storagev1.PrepareUploadResponse{
		UploadToken: uploadToken,
		StoragePath: generateStoragePath(fileID, req.Filename),
		BaseResponse: &sharedv1.Response{
			Message: "Upload prepared successfully",
		},
	}, nil
}

// CompleteUpload finalizes the file upload process
func (s *FileStorageService) CompleteUpload(
	ctx context.Context,
	req *storagev1.CompleteUploadRequest,
) (*storagev1.CompleteUploadResponse, error) {
	// Validate input
	if req == nil || req.FileMetadata == nil {
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

func generateStoragePath(fileId string, fileName string) string {
	now := time.Now()
	return filepath.Join(
		"uploads",
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		fmt.Sprintf("%s_%s", fileId, fileName),
	)
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
