package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
)

var (
	ErrFileNotFound = errors.New("file not found")
)

// StorageRepository defines the interface for file storage operations
type StorageRepository interface {
	// Store saves file metadata
	Store(ctx context.Context, storage *models.Storage) error

	// FindByID retrieves file metadata by ID
	FindByID(ctx context.Context, fileID string) (*models.Storage, error)

	// List retrieves files with pagination
	List(ctx context.Context, userID string, page, pageSize int) ([]*models.Storage, int, error)

	// Delete removes a file by ID
	Delete(ctx context.Context, fileID, userID string) error
}

// SQLiteStorageRepository implements StorageRepository for SQLite
type SQLiteStorageRepository struct {
	db *sql.DB
}

// NewSQLiteStorageRepository creates a new SQLite-based repository
func NewSQLiteStorageRepository(db *sql.DB) *SQLiteStorageRepository {
	return &SQLiteStorageRepository{db: db}
}

// Implement repository methods...
