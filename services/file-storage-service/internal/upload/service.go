package upload

import (
	"context"
	"time"

	file "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/file"
	repository "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata"
	storage "github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage"
	validation "github.com/yaanno/upload-store-process/services/file-storage-service/internal/upload/validation"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"google.golang.org/grpc/codes"
)

type UploadService interface {
	Upload(context.Context, *UploadRequest) (*UploadResponse, error)
}

type UploadServiceImpl struct {
	metadataRepo repository.FileMetadataRepository
	storage      storage.Provider
	logger       *logger.Logger
}

func NewUploadService(
	metadataRepo repository.FileMetadataRepository,
	storage storage.Provider,
	logger *logger.Logger,
) *UploadServiceImpl {
	return &UploadServiceImpl{
		metadataRepo: metadataRepo,
		storage:      storage,
		logger:       logger,
	}
}

func (s *UploadServiceImpl) Upload(ctx context.Context, req *UploadRequest) (*UploadResponse, error) {
	// Validate input
	if err := validation.ValidateSecureUploadToken(req.StorageUploadToken, req.FileID); err != nil {
		s.logger.Error().Err(err).Msg("invalid upload token")
		return nil, &UploadError{
			Code:    codes.PermissionDenied,
			Message: "invalid upload token",
			Err:     err,
		}
	}

	// Retrieve file metadata
	metadata, err := s.metadataRepo.RetrieveFileMetadataByID(ctx, req.FileID)
	if err != nil {
		s.logger.Error().Err(err).Str("fileId", req.FileID).Msg("failed to retrieve file metadata")
		return nil, &UploadError{
			Code:    codes.NotFound,
			Message: "file metadata not found",
			Err:     err,
		}
	}

	// Validate file state
	if metadata.ProcessingStatus != string(file.StatusPending) {
		return nil, &UploadError{
			Code:    codes.FailedPrecondition,
			Message: "invalid file upload state",
			Err:     nil,
		}
	}

	// Store the file
	storagePath, err := s.storage.Store(ctx, req.FileID, req.FileContent)
	if err != nil {
		metadata.ProcessingStatus = string(file.StatusPending)
		_ = s.metadataRepo.UpdateFileMetadata(ctx, metadata)
		return nil, &UploadError{
			Code:    codes.Internal,
			Message: "failed to store file",
			Err:     err,
		}
	}

	// Update metadata
	metadata.ProcessingStatus = string(file.StatusComplete)
	metadata.StoragePath = storagePath
	metadata.UpdatedAt = time.Now().UTC()

	if err := s.metadataRepo.UpdateFileMetadata(ctx, metadata); err != nil {
		s.logger.Error().Err(err).Str("fileID", metadata.ID).Msg("Failed to update file metadata")
	}

	return &UploadResponse{
		FileID:      metadata.ID,
		StoragePath: storagePath,
		Message:     "File uploaded successfully",
	}, nil
}

var _ UploadService = (*UploadServiceImpl)(nil)
