package upload

import (
	"context"
	"time"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	uploadv1 "github.com/yaanno/upload-store-process/gen/go/fileupload/v1"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	file "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/file"
	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	repository "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata"
	storage "github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage"
	token "github.com/yaanno/upload-store-process/services/file-storage-service/internal/upload/token"
	validation "github.com/yaanno/upload-store-process/services/file-storage-service/internal/upload/validation"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UploadService interface {
	Upload(context.Context, *UploadRequest) (*UploadResponse, error)
	PrepareUpload(context.Context, *storagev1.PrepareUploadRequest) (*storagev1.PrepareUploadResponse, error)
}

type UploadServiceImpl struct {
	metadataRepo repository.FileMetadataRepository
	storage      storage.Provider
	logger       logger.Logger
}

func NewUploadService(
	metadataRepo repository.FileMetadataRepository,
	storage storage.Provider,
	logger logger.Logger,
) *UploadServiceImpl {
	return &UploadServiceImpl{
		metadataRepo: metadataRepo,
		storage:      storage,
		logger:       logger,
	}
}

// PrepareUpload prepares a file upload by storing initial metadata
func (s *UploadServiceImpl) PrepareUpload(
	ctx context.Context,
	req *storagev1.PrepareUploadRequest,
) (*storagev1.PrepareUploadResponse, error) {
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
	storagePath := s.storage.GenerateStoragePath(fileID)

	// Prepare initial metadata
	initialMetadata := &sharedv1.FileMetadata{
		FileId:           fileID,
		OriginalFilename: req.Filename,
		FileSizeBytes:    req.FileSizeBytes,
		FileType:         file.DetermineFileType(req.Filename),
		UploadTimestamp:  time.Now().Unix(),
		UserId:           req.UserId,
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

	return &storagev1.PrepareUploadResponse{
		StorageUploadToken: uploadToken,
		StoragePath:        storagePath,
		GlobalUploadId:     fileID,
		BaseResponse: &sharedv1.Response{
			Message: "Upload prepared successfully",
		},
	}, nil
}

// func (s *UploadServiceImpl) UploadHttp(ctx context.Context, req *UploadFileRequest) error {
// 	if err := s.metadataRepo.CreateFileMetadata(ctx, &domain.FileMetadataRecord{}); err != nil {
// 		return err
// 	}
// 	_, err := s.storage.Store(ctx, req.FileId, req.FileContent)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// Update interfa
// Replace both UploadHttp and UploadRPC with a single Upload method
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

// UploadFile handles the actual file upload process
// func (s *UploadServiceImpl) UploadRPC(
// 	ctx context.Context,
// 	req *uploadv1.UploadFileRequest,
// ) (*uploadv1.UploadFileResponse, error) {
// 	// Validate input
// 	if err := s.ValidateUploadFileRequest(ctx, req); err != nil {
// 		return nil, err
// 	}

// 	// Retrieve file metadata to confirm upload context
// 	metadata, err := s.metadataRepo.RetrieveFileMetadataByID(ctx, req.FileId)
// 	if err != nil {
// 		s.logger.Error().
// 			Str("method", "UploadFile").
// 			Err(err).
// 			Str("fileId", req.FileId).
// 			Msg("failed to retrieve file metadata")
// 		return nil, status.Errorf(codes.NotFound, "file metadata not found")
// 	}

// 	// Validate that file is in a valid state for upload
// 	if metadata.ProcessingStatus != "PENDING" {
// 		s.logger.Error().
// 			Str("method", "UploadFile").
// 			Str("fileId", metadata.ID).
// 			Str("processingStatus", metadata.ProcessingStatus).
// 			Msg("invalid file upload state")
// 		return nil, status.Errorf(codes.FailedPrecondition, "invalid file upload state")
// 	}

// 	// Store the file using the storage provider
// 	storagePath, err := s.storage.Store(
// 		ctx,
// 		req.FileId,
// 		bytes.NewReader(req.FileContent),
// 	)
// 	if err != nil {
// 		// Rollback metadata status
// 		metadata.ProcessingStatus = string(file.StatusPending)
// 		if err := s.metadataRepo.UpdateFileMetadata(ctx, metadata); err != nil {
// 			s.logger.Error().
// 				Str("method", "UploadFile").
// 				Err(err).
// 				Str("fileId", metadata.ID).
// 				Msg("failed to rollback file metadata")
// 		}
// 		return nil, &UploadError{
// 			Code:    codes.Internal,
// 			Message: "failed to store file",
// 			Err:     err,
// 		}
// 	}

// 	// Update metadata to COMPLETED status
// 	metadata.ProcessingStatus = string(file.StatusComplete)
// 	metadata.StoragePath = storagePath
// 	metadata.UpdatedAt = time.Now().UTC()

// 	if err := s.metadataRepo.UpdateFileMetadata(ctx, metadata); err != nil {
// 		s.logger.Error().
// 			Str("method", "UploadFile").
// 			Str("fileID", metadata.ID).
// 			Err(err).
// 			Msg("Failed to update file metadata after upload")
// 		// Non-critical error, file is already stored
// 	}

// 	return &uploadv1.UploadFileResponse{
// 		BaseResponse: &sharedv1.Response{
// 			Message: "File uploaded successfully",
// 		},
// 		StoragePath: storagePath,
// 		FileId:      metadata.ID,
// 	}, nil
// }

// ValidateUploadRequest validates the upload request
func (s *UploadServiceImpl) ValidateUploadFileRequest(ctx context.Context, req *uploadv1.UploadFileRequest) error {

	// Validate upload token
	if err := validation.ValidateSecureUploadToken(req.StorageUploadToken, req.FileId); err != nil {
		s.logger.Error().
			Str("method", "UploadFile").
			Err(err).
			Msg("invalid upload token")
		return status.Errorf(codes.PermissionDenied, "invalid upload token")
	}
	return nil
}

var _ UploadService = (*UploadServiceImpl)(nil)
