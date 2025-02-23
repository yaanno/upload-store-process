package metadata

import (
	"context"
	"fmt"
	"time"

	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/file"
	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	token "github.com/yaanno/upload-store-process/services/file-storage-service/internal/upload/token"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PrepareUploadParams struct {
	FileName string
	FileSize int64
	UserID   string
}

type PrepareUploadResult struct {
	FileID      string
	StoragePath string
	UploadToken string
	ExpiresAt   time.Time
	Message     string
}

// MetadataService defines the interface for file metadata operations
type MetadataService interface {
	CreateFileMetadata(ctx context.Context) error
	GetFileMetadata(ctx context.Context, userID string, fileID string) (*domain.FileMetadataRecord, error)
	DeleteFileMetadata(ctx context.Context, userID string, fileID string) error
	ListFileMetadata(ctx context.Context, opts *domain.FileMetadataListOptions) (records []*domain.FileMetadataRecord, err error)
	PrepareUpload(ctx context.Context, params *PrepareUploadParams) (*PrepareUploadResult, error)
	CleanupExpiredMetadata(ctx context.Context) (int64, error)
	UpdateFileMetadata(ctx context.Context, fileID string, record *domain.FileMetadataRecord) error
	RetrieveFileMetadataByID(ctx context.Context, fileID string) (*domain.FileMetadataRecord, error)
}

type MetadataServiceImpl struct {
	metadataRepo FileMetadataRepository
	logger       *logger.Logger
}

// NewMetadataService creates a new metadata service
func NewMetadataService(metadataRepo FileMetadataRepository, logger *logger.Logger) *MetadataServiceImpl {
	return &MetadataServiceImpl{
		metadataRepo: metadataRepo,
		logger:       logger,
	}
}

func (s *MetadataServiceImpl) UpdateFileMetadata(ctx context.Context, fileID string, record *domain.FileMetadataRecord) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := s.metadataRepo.UpdateFileMetadata(ctx, record); err != nil {
		s.logger.Error().
			Str("method", "UpdateFileMetadata").
			Err(err).
			Str("fileId", fileID).
			Msg("failed to update file metadata")
		return err
	}
	return nil
}

func (s *MetadataServiceImpl) CreateFileMetadata(ctx context.Context) error {
	// cancels the context if the request is canceled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// create metadata record structure, sample for now
	metadata := &domain.FileMetadataRecord{
		ID:               "123",
		StoragePath:      "test",
		ProcessingStatus: "test",
	}
	// call repository operation
	if err := s.metadataRepo.CreateFileMetadata(ctx, metadata); err != nil { // log error
		s.logger.Error().
			Str("method", "CreateFileMetadata").
			Err(err).
			Msg("failed to create file metadata")
		// return error
		return err
	}
	// return nil
	return nil
}

func (s *MetadataServiceImpl) GetFileMetadata(ctx context.Context, userID string, fileID string) (record *domain.FileMetadataRecord, err error) {
	// cancels the context if the request is canceled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := s.validateFileOwnership(ctx, userID, fileID); err != nil {
		// log error
		s.logger.Error().
			Str("method", "DeleteFileMetadata").
			Err(err).
			Str("fileId", fileID).
			Msg("failed to validate file ownership")
		// return error
		return nil, status.Errorf(codes.PermissionDenied, "failed to validate file ownership")
	}

	// call repository operation
	record, err = s.metadataRepo.RetrieveFileMetadataByID(ctx, fileID)
	if err != nil {
		// log error
		s.logger.Error().
			Str("method", "GetFileMetadata").
			Err(err).
			Str("fileId", fileID).
			Msg("failed to retrieve file metadata")
		// return error
		return nil, err
	}
	// return record
	return record, nil
}

func (s *MetadataServiceImpl) DeleteFileMetadata(ctx context.Context, userID string, fileID string) error {
	// cancels the context if the request is canceled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// validate ownership
	if err := s.validateFileOwnership(ctx, userID, fileID); err != nil {
		// log error
		s.logger.Error().
			Str("method", "DeleteFileMetadata").
			Err(err).
			Str("fileId", fileID).
			Msg("failed to validate file ownership")
		// return error
		return status.Errorf(codes.PermissionDenied, "failed to validate file ownership")
	}

	if err := s.metadataRepo.RemoveFileMetadata(ctx, fileID); err != nil {
		// log error
		s.logger.Error().
			Str("method", "DeleteFileMetadata").
			Err(err).
			Str("fileId", fileID).
			Msg("failed to delete file metadata")
		// return error
		return err
	}
	return nil
}

func (s *MetadataServiceImpl) ListFileMetadata(ctx context.Context, opts *domain.FileMetadataListOptions) (records []*domain.FileMetadataRecord, err error) {
	// cancels the context if the request is canceled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts.Status = string(file.StatusComplete)

	// call repository operation
	records, err = s.metadataRepo.ListFileMetadata(ctx, opts)
	if err != nil {
		// log error
		s.logger.Error().
			Str("method", "ListFileMetadata").
			Err(err).
			Msg("failed to list file metadata")
		// return error
		return nil, err
	}
	if records == nil {
		// log error
		s.logger.Error().
			Str("method", "ListFileMetadata").
			Err(err).
			Msg("no file metadata found")
		// return error
		return nil, status.Errorf(codes.NotFound, "no file metadata found")
	}
	// return records
	return records, nil
}

func (s *MetadataServiceImpl) PrepareUpload(ctx context.Context, params *PrepareUploadParams) (*PrepareUploadResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Generate secure file ID
	fileID, err := token.GenerateSecureFileID()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate file ID")
		return nil, status.Errorf(codes.Internal, "failed to generate file ID")
	}

	// Generate upload token
	uploadToken, err := token.GenerateSecureUploadToken(fileID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate upload token")
		return nil, status.Errorf(codes.Internal, "failed to generate upload token")
	}

	// Generate storage path using the storage provider
	storagePath, err := s.generateStoragePath(fileID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate storage path")
		return nil, status.Errorf(codes.Internal, "failed to generate storage path")
	}

	// Prepare initial metadata
	initialMetadata := &sharedv1.FileMetadata{
		FileId:           fileID,
		OriginalFilename: params.FileName,
		FileSizeBytes:    params.FileSize,
		ContentType:      file.DetermineFileType(params.FileName),
		CreatedAt:        timestamppb.Now(),
		UserId:           params.UserID,
	}

	// Create initial metadata record
	metadataRecord := &domain.FileMetadataRecord{
		ID:               fileID,
		Metadata:         initialMetadata,
		StoragePath:      storagePath,
		ProcessingStatus: string(file.StatusPending),
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	// Store initial metadata
	if err := s.metadataRepo.CreateFileMetadata(ctx, metadataRecord); err != nil {
		s.logger.Error().Err(err).Msg("Failed to create initial file metadata")
		return nil, status.Errorf(codes.Internal, "failed to create file metadata")
	}

	return &PrepareUploadResult{
		FileID:      fileID,
		UploadToken: uploadToken,
		StoragePath: storagePath,
		ExpiresAt:   time.Now().Add(time.Hour * 1),
		Message:     "File upload prepared successfully",
	}, nil
}

func (s *MetadataServiceImpl) CleanupExpiredMetadata(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	expirationThreshold := time.Now().Add(-24 * time.Hour)
	// Delete expired metadata records
	result, err := s.metadataRepo.CleanupExpiredMetadata(ctx, expirationThreshold)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to delete expired file metadata")
		return 0, status.Errorf(codes.Internal, "failed to delete expired file metadata")
	}
	return result, nil
}

func (s *MetadataServiceImpl) RetrieveFileMetadataByID(ctx context.Context, fileID string) (*domain.FileMetadataRecord, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// call repository operation
	record, err := s.metadataRepo.RetrieveFileMetadataByID(ctx, fileID)
	if err != nil {
		// log error
		s.logger.Error().
			Str("method", "RetrieveFileMetadataByID").
			Err(err).
			Str("fileId", fileID).
			Msg("failed to retrieve file metadata")
		// return error
		return nil, err
	}
	// return record
	return record, nil
}

var _ MetadataService = (*MetadataServiceImpl)(nil)

// ValidateGetFileMetadataRequest validates the get file metadata request
func (s *MetadataServiceImpl) validateFileOwnership(ctx context.Context, userID string, fileID string) error {

	opts := &domain.FileMetadataListOptions{
		UserID: userID,
		FileID: fileID,
	}

	isOwner, err := s.metadataRepo.IsFileOwnedByUser(ctx, opts)
	if err != nil {
		s.logger.Error().
			Str("method", "GetFileMetadata").
			Err(err).
			Str("file_id", fileID).
			Msg("failed to check file ownership")
		return status.Errorf(codes.Internal, "failed to check file ownership: %v", err)
	}

	if !isOwner {
		s.logger.Warn().
			Str("method", "GetFileMetadata").
			Str("file_id", fileID).
			Str("user_id", userID).
			Msg("user does not own file")
		return status.Errorf(codes.PermissionDenied, "user does not own file")
	}

	return nil
}

func (s *MetadataServiceImpl) generateStoragePath(fileID string) (string, error) {
	// Implement storage path generation logic here
	// This could involve combining the fileID with a timestamp or other factors
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d-%s", timestamp, fileID), nil
}
