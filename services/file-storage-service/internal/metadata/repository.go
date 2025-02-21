package metadata

import (
	"context"
	"errors"

	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	sqliteRepositoy "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata/implementations/sqlite"
)

// FileMetadataRepository defines the interface for file metadata storage operations
type FileMetadataRepository interface {
	CreateFileMetadata(ctx context.Context, metadata *domain.FileMetadataRecord) error
	RetrieveFileMetadataByID(ctx context.Context, fileID string) (*domain.FileMetadataRecord, error)
	ListFileMetadata(ctx context.Context, opts *domain.FileMetadataListOptions) ([]*domain.FileMetadataRecord, error)
	ListFiles(ctx context.Context, opts *domain.FileMetadataListOptions) ([]*domain.FileMetadataRecord, int, error)
	RemoveFileMetadata(ctx context.Context, fileID string) error
	UpdateFileMetadata(ctx context.Context, metadata *domain.FileMetadataRecord) error
	IsFileOwnedByUser(ctx context.Context, opts *domain.FileMetadataListOptions) (bool, error)
	SoftDeleteFile(ctx context.Context, fileID, userID string) error
}

type RepositoryType string

const (
	SQLite RepositoryType = "sqlite"
)

func NewRepository(repoType RepositoryType, cfg interface{}) (FileMetadataRepository, error) {
	switch repoType {
	case SQLite:
		return sqliteRepositoy.NewSQLiteFileMetadataRepository(), nil
	default:
		return nil, errors.New("invalid repository type")
	}
}

var _ FileMetadataRepository = (*sqliteRepositoy.SQLiteFileMetadataRepository)(nil)
