package metadata

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	sqliteRepository "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata/implementations/sqlite"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

// FileMetadataRepository defines the interface for file metadata storage operations
type FileMetadataRepository interface {
	CreateFileMetadata(ctx context.Context, metadata *domain.FileMetadataRecord) error
	RetrieveFileMetadataByID(ctx context.Context, fileID string) (*domain.FileMetadataRecord, error)
	ListFileMetadata(ctx context.Context, opts *domain.FileMetadataListOptions) ([]*domain.FileMetadataRecord, error)
	RemoveFileMetadata(ctx context.Context, fileID string) error
	UpdateFileMetadata(ctx context.Context, metadata *domain.FileMetadataRecord) error
	IsFileOwnedByUser(ctx context.Context, opts *domain.FileMetadataListOptions) (bool, error)
	SoftDeleteMetadata(ctx context.Context, fileID, userID string) error
	CleanupExpiredMetadata(ctx context.Context, expirationTime time.Time) (int64, error)
}

type RepositoryType string

const (
	SQLite RepositoryType = "sqlite"
)

func NewRepository(repoType RepositoryType, db interface{}, logger *logger.Logger) (FileMetadataRepository, error) {
	switch repoType {
	case SQLite:
		sqlDb, ok := db.(*sql.DB)
		if !ok {
			return nil, errors.New("invalid database type")
		}
		return sqliteRepository.NewSQLiteFileMetadataRepository(sqlDb, logger), nil
	default:
		return nil, errors.New("invalid repository type")
	}
}

var _ FileMetadataRepository = (*sqliteRepository.SQLiteFileMetadataRepository)(nil)
