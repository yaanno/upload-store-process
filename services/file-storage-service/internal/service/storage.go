package service

import (
	"context"
	"time"

	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/repository"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
)

// FileStorageService implements the gRPC service
type FileStorageService struct {
	storagev1.UnimplementedFileStorageServiceServer
	repo   repository.FileMetadataRepository
	logger *logger.Logger
}

// NewFileStorageService creates a new service instance
func NewFileStorageService(
	repo repository.FileMetadataRepository,
	logger *logger.Logger,
) *FileStorageService {
	return &FileStorageService{
		repo:   repo,
		logger: logger,
	}
}

// StoreFileMetadata implements the gRPC method
func (s *FileStorageService) StoreFileMetadata(
	ctx context.Context,
	req *storagev1.StoreFileMetadataRequest,
) (*storagev1.StoreFileMetadataResponse, error) {
	// Validate input
	if req.Metadata == nil {
		return nil, status.Errorf(codes.InvalidArgument, "metadata is required")
	}

	// Convert to internal model
	storageModel := &models.FileMetadataRecord{
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
	}

	// Store metadata
	if err := s.repo.CreateFileMetadata(ctx, storageModel); err != nil {
		s.logger.Error().Err(err).Msg("Failed to store file metadata")
		return nil, status.Errorf(codes.Internal, "failed to store metadata")
	}

	return &storagev1.StoreFileMetadataResponse{
		BaseResponse: &sharedv1.Response{
			Code:    sharedv1.Response_STATUS_CODE_OK_UNSPECIFIED,
			Message: "Metadata stored successfully",
		},
		FileId: storageModel.ID,
	}, nil
}

// Implement other gRPC methods: GetFileMetadata, ListFiles, DeleteFile, etc.
