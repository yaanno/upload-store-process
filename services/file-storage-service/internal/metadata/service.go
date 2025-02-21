package metadata

// import (
// 	"context"

// 	"github.com/rs/zerolog"
// 	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
// 	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
// 	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
// 	metadataRepo "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata/repository/sqlite"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// type MetadataService interface {
// 	CreateFileMetadata(ctx context.Context, req *storagev1.) (*storagev1.CreateFileMetadataResponse, error)
// 	GetFileMetadata(ctx context.Context, req *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error)
// }

// type MetadataServiceImpl struct {
// 	metadataRepo metadataRepo.SQLiteFileMetadataRepository
// 	logger       *zerolog.Logger
// }

// // NewMetadataService creates a new metadata service
// func NewMetadataService(metadataRepo metadataRepo.SQLiteFileMetadataRepository, logger *zerolog.Logger) *MetadataServiceImpl {
// 	return &MetadataServiceImpl{
// 		metadataRepo: metadataRepo,
// 		logger:       logger,
// 	}
// }

// // GetFileMetadata implements v1.FileStorageServiceServer.
// func (s *MetadataServiceImpl) GetFileMetadata(ctx context.Context, req *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error) {

// 	// Retrieve file metadata
// 	metadata, err := s.repo.RetrieveFileMetadataByID(ctx, req.FileId)
// 	if err != nil {
// 		s.logger.Error().
// 			Str("method", "GetFileMetadata").
// 			Err(err).
// 			Str("fileId", req.FileId).
// 			Msg("failed to retrieve file metadata")
// 		return nil, status.Errorf(codes.NotFound, "file metadata not found")
// 	}

// 	if err := s.ValidateGetFileMetadataRequest(ctx, req); err != nil {
// 		return nil, err
// 	}

// 	// Transform metadata to GetFileMetadataResponse
// 	fileMetadata := &sharedv1.FileMetadata{
// 		FileId:           metadata.ID,
// 		OriginalFilename: metadata.Metadata.OriginalFilename,
// 		FileSizeBytes:    metadata.Metadata.FileSizeBytes,
// 		FileType:         metadata.Metadata.FileType,
// 		UploadTimestamp:  metadata.Metadata.UploadTimestamp,
// 		UserId:           metadata.Metadata.UserId,
// 		StoragePath:      metadata.StoragePath,
// 	}

// 	// Return response
// 	return &storagev1.GetFileMetadataResponse{
// 		BaseResponse: &sharedv1.Response{
// 			Message: "File metadata retrieved successfully",
// 		},
// 		Metadata: fileMetadata,
// 	}, nil
// }

// // ValidateGetFileMetadataRequest validates the get file metadata request
// func (s *MetadataServiceImpl) ValidateGetFileMetadataRequest(ctx context.Context, req *storagev1.GetFileMetadataRequest) error {

// 	opts := &domain.FileMetadataListOptions{
// 		UserID: req.UserId,
// 		FileID: req.FileId,
// 	}

// 	isOwner, err := s.repo.IsFileOwnedByUser(ctx, opts)
// 	if err != nil {
// 		s.logger.Error().
// 			Str("method", "GetFileMetadata").
// 			Err(err).
// 			Str("file_id", req.FileId).
// 			Msg("failed to check file ownership")
// 		return status.Errorf(codes.Internal, "failed to check file ownership: %v", err)
// 	}

// 	if !isOwner {
// 		s.logger.Warn().
// 			Str("method", "GetFileMetadata").
// 			Str("file_id", req.FileId).
// 			Str("user_id", req.UserId).
// 			Msg("user does not own file")
// 		return status.Errorf(codes.PermissionDenied, "user does not own file")
// 	}

// 	return nil
// }
