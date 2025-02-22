package metadata

import (
	"context"
	"time"

	"github.com/google/uuid"
	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	GetFileMetadata(ctx context.Context, fileID string) (*domain.FileMetadataRecord, error)
	DeleteFileMetadata(ctx context.Context, fileID string) error
	ListFileMetadata(ctx context.Context, opts *domain.FileMetadataListOptions) (records []*domain.FileMetadataRecord, err error)
	PrepareUpload(ctx context.Context, params *PrepareUploadParams) (*PrepareUploadResult, error)
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

func (s *MetadataServiceImpl) GetFileMetadata(ctx context.Context, fileID string) (record *domain.FileMetadataRecord, err error) {
	// cancels the context if the request is canceled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var userID string // TODO: get user id from context
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

func (s *MetadataServiceImpl) DeleteFileMetadata(ctx context.Context, fileID string) error {
	// cancels the context if the request is canceled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// validate ownership
	var userID string // TODO: get user id from context
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

	metadata := &domain.FileMetadataRecord{
		ID: uuid.New().String(),
		// FileName:  params.FileName,
		// FileSize:  params.FileSize,
		// UserID:    params.UserID,
		// Status:    domain.,
		CreatedAt: time.Now(),
	}

	if err := s.metadataRepo.CreateFileMetadata(ctx, metadata); err != nil {
		s.logger.Error().
			Str("method", "PrepareUpload").
			Err(err).
			Msg("failed to create file metadata")
		return nil, err
	}

	return &PrepareUploadResult{
		FileID:      metadata.ID,
		StoragePath: metadata.StoragePath,
		// UploadToken: generateUploadToken(metadata.ID),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
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
