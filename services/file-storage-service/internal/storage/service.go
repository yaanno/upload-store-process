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
}

type StorageServiceImpl struct {
	logger   logger.Logger
	provider Provider
}

func NewStorageService(
	logger logger.Logger,
	provider Provider,
) *StorageServiceImpl {
	return &StorageServiceImpl{
		logger:   logger,
		provider: provider,
	}
}

func (s *StorageServiceImpl) Store(ctx context.Context, fileID string, content io.Reader) (string, error) {
	return s.provider.Store(ctx, fileID, content)
}

func (s *StorageServiceImpl) List(ctx context.Context) ([]string, error) {
	return s.provider.List(ctx)
}

func (s *StorageServiceImpl) Delete(ctx context.Context, fileID string) error {
	return s.provider.Delete(ctx, fileID)
}

func (s *StorageServiceImpl) Retrieve(ctx context.Context, fileID string) (io.ReadCloser, error) {
	return s.provider.Retrieve(ctx, fileID)
}

var _ StorageService = &StorageServiceImpl{}
