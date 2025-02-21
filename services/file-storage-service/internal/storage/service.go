package storage

import (
	"context"
	"io"

	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

type StorageService interface {
	Store(ctx context.Context, fileID string, content io.Reader) (string, error)
	Retrieve(ctx context.Context, fileID string) (io.ReadCloser, error)
	Delete(ctx context.Context, fileID string) error
	List(ctx context.Context) ([]string, error)
	GenerateStoragePath(fileID string) string
}

// FileStorageService implements the gRPC service
type StorageServiceImpl struct {
	logger   logger.Logger
	provider Provider
}

// NewFileStorageService creates a new instance of FileStorageService
func NewStorageService(
	logger logger.Logger,
	provider Provider,
) *StorageServiceImpl {
	return &StorageServiceImpl{
		logger:   logger,
		provider: provider,
	}
}

func (s *StorageServiceImpl) GenerateStoragePath(fileID string) string {
	return s.provider.GenerateStoragePath(fileID)
}

// StoreFile implements v1.FileStorageServiceServer.
func (s *StorageServiceImpl) Store(ctx context.Context, fileID string, content io.Reader) (string, error) {
	return s.provider.Store(ctx, fileID, content)

}

func (s *StorageServiceImpl) List(ctx context.Context) ([]string, error) {
	return s.provider.List(ctx)
}

// ListFiles retrieves a list of files for a user
// func (s *FileStorageServiceImpl) List(ctx context.Context, req *storagev1.ListFilesRequest) (*storagev1.ListFilesResponse, error) {
// 	// Calculate pagination
// 	pageSize := int32(req.PageSize)
// 	pageNum := int32(req.Page)

// 	// Prepare list options
// 	listOpts := &domain.FileMetadataListOptions{
// 		UserID: req.UserId,
// 		Limit:  int(pageSize),
// 		Offset: int((pageNum - 1) * pageSize),
// 	}

// 	// Retrieve file metadata
// 	fileMetadataList, totalCount, err := s.repo.ListFiles(ctx, listOpts)
// 	if err != nil {
// 		s.logger.Error().Err(err).Msg("Failed to list file metadata")
// 		return nil, status.Errorf(codes.Internal, "failed to list files")
// 	}

// 	// Convert to gRPC response
// 	var files []*sharedv1.FileMetadata
// 	for _, metadata := range fileMetadataList {
// 		files = append(files, &sharedv1.FileMetadata{
// 			FileId:           metadata.ID,
// 			OriginalFilename: metadata.Metadata.OriginalFilename,
// 			FileSizeBytes:    metadata.Metadata.FileSizeBytes,
// 			UploadTimestamp:  metadata.Metadata.UploadTimestamp,
// 		})
// 	}

// 	// Calculate total pages
// 	var totalPages int32
// 	if pageSize > 0 {
// 		totalPages = int32(math.Ceil(float64(totalCount) / float64(pageSize)))
// 	}

// 	return &storagev1.ListFilesResponse{
// 		Files:      files,
// 		TotalFiles: int32(totalCount),
// 		TotalPages: totalPages,
// 	}, nil
// }

// DeleteFile implements v1.FileStorageServiceServer.
func (s *StorageServiceImpl) Delete(ctx context.Context, fileID string) error {
	return s.provider.Delete(ctx, fileID)

	// if err := s.ValidateDeleteFileRequest(ctx, req); err != nil {
	// 	return nil, err
	// }

	// //  delete file from database
	// if err := s.repo.SoftDeleteFile(ctx, req.FileId, req.UserId); err != nil {
	// 	s.logger.Error().
	// 		Str("method", "DeleteFile").
	// 		Err(err).
	// 		Str("fileId", req.FileId).
	// 		Msg("failed to soft delete file metadata")
	// 	return nil, status.Errorf(codes.Internal, "failed to soft delete file metadata: %v", err)
	// }

	// return &storagev1.DeleteFileResponse{
	// 	FileDeleted: true,
	// 	DeletedAt:   timestamppb.Now(),
	// 	BaseResponse: &sharedv1.Response{
	// 		Message: "File deleted successfully",
	// 	},
	// }, nil
}

func (s *StorageServiceImpl) Retrieve(ctx context.Context, fileID string) (io.ReadCloser, error) {
	return s.provider.Retrieve(ctx, fileID)
}

// Internal Request validators

// ValidateDeleteRequest validates the delete request
// func (s *StorageServiceImpl) ValidateDeleteFileRequest(ctx context.Context, req *storagev1.DeleteFileRequest) error {

// 	opts := &domain.FileMetadataListOptions{
// 		UserID: req.UserId,
// 		FileID: req.FileId,
// 	}

// 	isOwner, err := s.repo.IsFileOwnedByUser(ctx, opts)
// 	if err != nil {
// 		s.logger.Error().
// 			Str("method", "DeleteFile").
// 			Err(err).
// 			Str("fileId", req.FileId).
// 			Msg("failed to check file ownership")
// 		return status.Errorf(codes.Internal, "failed to check file ownership: %v", err)
// 	}

// 	if !isOwner && !req.ForceDelete {
// 		s.logger.Warn().
// 			Str("method", "DeleteFile").
// 			Str("fileId", req.FileId).
// 			Str("userId", req.UserId).
// 			Msg("user does not own file")
// 		return status.Errorf(codes.PermissionDenied, "user does not own file")
// 	}

// 	return nil
// }

var _ StorageService = &StorageServiceImpl{}
