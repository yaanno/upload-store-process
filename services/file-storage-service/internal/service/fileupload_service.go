package service

import (
	"context"
	"io"

	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/repository"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage"
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
	storage    storage.FileStorageProvider
}

func NewFileUploadServiceImpl(
	repository repository.FileMetadataRepository,
	storage storage.FileStorageProvider,
) *FileUploadServiceImpl {
	return &FileUploadServiceImpl{
		repository: repository,
		storage:    storage,
	}
}

func (s *FileUploadServiceImpl) UploadFile(ctx context.Context, req *UploadFileRequest) error {
	if err := s.repository.CreateFileMetadata(ctx, &models.FileMetadataRecord{}); err != nil {
		return err
	}
	_, err := s.storage.StoreFile(ctx, req.FileId, req.StorageUploadToken, req.FileContent)
	if err != nil {
		return err
	}
	return nil
}
