package repository

import (
	"context"

	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
	sqliteRepositoy "github.com/yaanno/upload-store-process/services/file-storage-service/internal/repository/sqlite"
)

// FileMetadataRepository defines the interface for file metadata storage operations
type FileMetadataRepository interface {
	CreateFileMetadata(ctx context.Context, metadata *models.FileMetadataRecord) error
	RetrieveFileMetadataByID(ctx context.Context, fileID string) (*models.FileMetadataRecord, error)
	ListFileMetadata(ctx context.Context, opts *models.FileMetadataListOptions) ([]*models.FileMetadataRecord, error)
	ListFiles(ctx context.Context, opts *models.FileMetadataListOptions) ([]*models.FileMetadataRecord, int, error)
	RemoveFileMetadata(ctx context.Context, fileID string) error
	UpdateFileMetadata(ctx context.Context, metadata *models.FileMetadataRecord) error
	IsFileOwnedByUser(ctx context.Context, opts *models.FileMetadataListOptions) (bool, error)
	SoftDeleteFile(ctx context.Context, fileID, userID string) error
}

var _ FileMetadataRepository = (*sqliteRepositoy.SQLiteFileMetadataRepository)(nil)
