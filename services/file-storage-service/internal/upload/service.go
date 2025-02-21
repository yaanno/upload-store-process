package upload

import (
	"context"
	"io"

	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	repository "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata"
	storageProvider "github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage/providers/local"
)

type UploadFileRequest struct {
	FileId             string
	StorageUploadToken string
	FileSizeBytes      string
	FileContent        io.Reader
}

type FileUploadService interface {
	UploadFile(context.Context, *UploadFileRequest) error
}

type FileUploadServiceImpl struct {
	repository repository.FileMetadataRepository
	storage    *storageProvider.LocalFileSystem
}

func NewFileUploadService(
	repository repository.FileMetadataRepository,
	storage *storageProvider.LocalFileSystem,
) *FileUploadServiceImpl {
	return &FileUploadServiceImpl{
		repository: repository,
		storage:    storage,
	}
}

func (s *FileUploadServiceImpl) UploadFile(ctx context.Context, req *UploadFileRequest) error {
	if err := s.repository.CreateFileMetadata(ctx, &domain.FileMetadataRecord{}); err != nil {
		return err
	}
	_, err := s.storage.Store(ctx, req.FileId, req.FileContent)
	if err != nil {
		return err
	}
	return nil
}

var _ FileUploadService = (*FileUploadServiceImpl)(nil)
