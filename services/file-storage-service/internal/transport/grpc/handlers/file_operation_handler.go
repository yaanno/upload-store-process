package handlers

import (
	"context"
	"math"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/upload"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FileOperationdHandler interface {
	ListFiles(ctx context.Context, req *storagev1.ListFilesRequest) (*storagev1.ListFilesResponse, error)
	DeleteFile(ctx context.Context, req *storagev1.DeleteFileRequest) (*storagev1.DeleteFileResponse, error)
	GetFileMetadata(ctx context.Context, req *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error)
}

type FileOperationdHandlerImpl struct {
	storagev1.UnimplementedFileStorageServiceServer
	metadataService metadata.MetadataService
	uploadService   upload.UploadService
	logger          *logger.Logger
}

func NewFileOperationdHandler(metadataService metadata.MetadataService, logger *logger.Logger) *FileOperationdHandlerImpl {
	return &FileOperationdHandlerImpl{
		metadataService: metadataService,
		logger:          logger,
	}
}

// ListFiles retrieves a list of files for a user
func (h *FileOperationdHandlerImpl) ListFiles(ctx context.Context, req *storagev1.ListFilesRequest) (*storagev1.ListFilesResponse, error) {
	// Calculate pagination
	pageSize := int32(req.PageSize)
	pageNum := int32(req.Page)

	// Prepare list options
	listOpts := domain.FileMetadataListOptions{
		UserID: req.UserId,
		Limit:  int(pageSize),
		Offset: int((pageNum - 1) * pageSize),
	}

	// Retrieve file metadata
	fileMetadataList, err := h.metadataService.ListFileMetadata(ctx, &listOpts)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list file metadata")
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

	totalCount := len(files)

	// Calculate total pages
	var totalPages int32
	if pageSize > 0 {
		totalPages = int32(math.Ceil(float64(totalCount) / float64(pageSize)))
	}
	return &storagev1.ListFilesResponse{
		Files:      files,
		TotalFiles: int32(totalCount),
		TotalPages: totalPages,
	}, nil
}

// DeleteFile implements v1.FileStorageServiceServer.
func (h *FileOperationdHandlerImpl) DeleteFile(ctx context.Context, req *storagev1.DeleteFileRequest) (*storagev1.DeleteFileResponse, error) {

	//  delete file from database
	if err := h.metadataService.DeleteFileMetadata(ctx, req.FileId); err != nil {
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
func (h *FileOperationdHandlerImpl) GetFileMetadata(ctx context.Context, req *storagev1.GetFileMetadataRequest) (*storagev1.GetFileMetadataResponse, error) {

	// Retrieve file metadata
	metadata, err := h.metadataService.GetFileMetadata(ctx, req.FileId)
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
		FileType:         metadata.Metadata.FileType,
		UploadTimestamp:  metadata.Metadata.UploadTimestamp,
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

func (h *FileOperationdHandlerImpl) PrepareUpload(ctx context.Context, req *storagev1.PrepareUploadRequest) (*storagev1.PrepareUploadResponse, error) {

	serviceReq := &upload.PrepareUploadRequest{
		Filename:      req.Filename,
		FileSizeBytes: req.FileSizeBytes,
		UserID:        req.UserId,
	}

	result, err := h.uploadService.PrepareUpload(ctx, serviceReq)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to prepare upload")
		return nil, status.Errorf(codes.Internal, "failed to prepare upload: %v", err)
	}

	return &storagev1.PrepareUploadResponse{
		StorageUploadToken: result.UploadToken,
		StoragePath:        result.StoragePath,
		GlobalUploadId:     result.FileID,
		BaseResponse: &sharedv1.Response{
			Message: result.Message,
		},
	}, nil
}

var _ FileOperationdHandler = (*FileOperationdHandlerImpl)(nil)
