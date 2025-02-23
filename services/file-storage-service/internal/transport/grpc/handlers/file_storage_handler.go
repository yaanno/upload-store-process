package handlers

import (
	"context"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FileStorageHandler interface {
	ListFiles(ctx context.Context, req *storagev1.ListFilesRequest) (*storagev1.ListFilesResponse, error)
	DeleteFile(ctx context.Context, req *storagev1.DeleteFileRequest) (*storagev1.DeleteFileResponse, error)
	GetFileMetadata(ctx context.Context, req *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error)
}

type FileStorageHandlerImpl struct {
	storagev1.UnimplementedFileStorageServiceServer
	metadataService metadata.MetadataService
	logger          *logger.Logger
}

func NewFileOperationdHandler(metadataService metadata.MetadataService, logger *logger.Logger) *FileStorageHandlerImpl {
	return &FileStorageHandlerImpl{
		metadataService: metadataService,
		logger:          logger,
	}
}

// ListFiles retrieves a list of files for a user
func (h *FileStorageHandlerImpl) ListFiles(ctx context.Context, req *storagev1.ListFilesRequest) (*storagev1.ListFilesResponse, error) {

	// Prepare list options
	listOpts := domain.FileMetadataListOptions{
		UserID: req.UserId,
	}

	// Retrieve file metadata
	fileMetadataList, err := h.metadataService.ListFileMetadata(ctx, &listOpts)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list file metadata")
		return nil, status.Errorf(codes.NotFound, "failed to list files")
	}

	if len(fileMetadataList) == 0 {
		h.logger.Info().Msg("No files found")
		return &storagev1.ListFilesResponse{
			TotalFiles: 0,
			BaseResponse: &sharedv1.Response{
				Message: "No files found",
			},
		}, nil
	}

	// Convert to gRPC response
	var files []*sharedv1.FileMetadata
	for _, metadata := range fileMetadataList {
		files = append(files, &sharedv1.FileMetadata{
			FileId:           metadata.ID,
			OriginalFilename: metadata.Metadata.OriginalFilename,
			FileSizeBytes:    metadata.Metadata.FileSizeBytes,
			CreatedAt:        metadata.Metadata.CreatedAt,
		})
	}

	totalCount := len(files)

	// Calculate total pages
	return &storagev1.ListFilesResponse{
		Files:      files,
		TotalFiles: int32(totalCount),
	}, nil
}

// DeleteFile implements v1.FileStorageServiceServer.
func (h *FileStorageHandlerImpl) DeleteFile(ctx context.Context, req *storagev1.DeleteFileRequest) (*storagev1.DeleteFileResponse, error) {

	//  delete file from database
	if err := h.metadataService.DeleteFileMetadata(ctx, req.UserId, req.FileId); err != nil {
		h.logger.Error().
			Str("method", "DeleteFile").
			Err(err).
			Str("fileId", req.FileId).
			Msg("failed to soft delete file metadata")
		return nil, status.Errorf(codes.Internal, "failed to delete file metadata: %v", err)
	}
	// TODO: delete file from storage provider

	return &storagev1.DeleteFileResponse{
		FileDeleted: true,
		DeletedAt:   timestamppb.Now(),
		BaseResponse: &sharedv1.Response{
			Message: "File deleted successfully",
		},
	}, nil
}

// // GetFileMetadata implements v1.FileStorageServiceServer.
func (h *FileStorageHandlerImpl) GetFileMetadata(ctx context.Context, req *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error) {
	// Retrieve file metadata
	metadata, err := h.metadataService.GetFileMetadata(ctx, req.UserId, req.FileId)
	if err != nil {
		h.logger.Error().
			Str("method", "GetFileMetadata").
			Err(err).
			Str("fileId", req.FileId).
			Msg("failed to retrieve file metadata")
		return nil, status.Errorf(codes.NotFound, "file metadata not found")
	}

	// Transform metadata to GetFileMetadataResponse
	fileMetadata := &sharedv1.FileMetadata{
		FileId:           metadata.ID,
		OriginalFilename: metadata.Metadata.OriginalFilename,
		FileSizeBytes:    metadata.Metadata.FileSizeBytes,
		ContentType:      metadata.Metadata.ContentType,
		CreatedAt:        metadata.Metadata.CreatedAt,
		UserId:           metadata.Metadata.UserId,
		StoragePath:      metadata.StoragePath,
	}

	// Return response
	return &storagev1.GetFileMetadataResponse{
		BaseResponse: &sharedv1.Response{
			Message: "File metadata retrieved successfully",
		},
		Metadata: fileMetadata,
	}, nil
}

func (h *FileStorageHandlerImpl) PrepareUpload(ctx context.Context, req *storagev1.PrepareUploadRequest) (*storagev1.PrepareUploadResponse, error) {
	h.logger.Info().Str("filename", req.Filename).Msg("Preparing upload")

	uploadParams := &metadata.PrepareUploadParams{
		FileName: req.Filename,
		FileSize: req.FileSizeBytes,
		UserID:   req.UserId,
	}

	result, err := h.metadataService.PrepareUpload(ctx, uploadParams)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to prepare upload")
		return nil, status.Errorf(codes.Internal, "failed to prepare upload: %v", err)
	}

	return &storagev1.PrepareUploadResponse{
		StorageUploadToken: result.UploadToken,
		StoragePath:        result.StoragePath,
		FileId:             result.FileID,
		ExpirationTime:     result.ExpiresAt.Unix(),
		BaseResponse: &sharedv1.Response{
			Message: result.Message,
		},
	}, nil
}

var _ FileStorageHandler = (*FileStorageHandlerImpl)(nil)
