package metadata

import (
	"context"
	"database/sql"
	"errors"

	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	sqliteRepositoy "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata/implementations/sqlite"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
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

func NewRepository(repoType RepositoryType, db interface{}, logger logger.Logger) (FileMetadataRepository, error) {
	switch repoType {
	case SQLite:
		sqlDb, ok := db.(*sql.DB)
		if !ok {
			return nil, errors.New("invalid database type")
		}
		return sqliteRepositoy.NewSQLiteFileMetadataRepository(sqlDb, logger), nil
	default:
		return nil, errors.New("invalid repository type")
	}
}

var _ FileMetadataRepository = (*sqliteRepositoy.SQLiteFileMetadataRepository)(nil)
