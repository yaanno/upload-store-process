package storage

import (
	"context"
	"errors"
	"io"

	metadataService "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata"
	filesystem "github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage/providers/local"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

// Provider defines the interface for file storage operations
type Provider interface {
	// Store saves a file and returns its storage path
	Store(ctx context.Context, fileID string, content io.Reader) (string, error)

	// Retrieve gets a file by its ID
	Retrieve(ctx context.Context, fileID string) (io.ReadCloser, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, fileID string) error

	// List returns a list of files in the storage
	List(ctx context.Context) ([]string, error)
}

type ProviderType string

type Config struct {
	BasePath        string `mapstructure:"base_path"`
	MetadataService metadataService.MetadataService
}

const (
	Local ProviderType = "local"
)

func NewProvider(providerType ProviderType, cfg interface{}, metadataService metadataService.MetadataService, logger *logger.Logger) (Provider, error) {
	switch providerType {
	case Local:
		localCfg, ok := cfg.(*Config)
		if !ok {
			return nil, errors.New("invalid configuration type")
		}
		return filesystem.NewLocalFileSystem(localCfg.BasePath, localCfg.MetadataService, logger), nil
	default:
		return nil, errors.New("invalid provider type")
	}
}
