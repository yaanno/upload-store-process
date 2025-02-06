package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
)

var (
	ErrFileNotFound = errors.New("file not found")
	ErrInvalidInput = errors.New("invalid input")
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

// Store saves file metadata with transaction support and upsert logic
func (r *SQLiteStorageRepository) Store(ctx context.Context, storage *models.Storage) error {
	// Validate input
	if storage == nil {
		return ErrInvalidInput
	}

	// Validate storage model
	if err := storage.Validate(); err != nil {
		return fmt.Errorf("invalid storage model: %v", err)
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Error rolling back transaction: %v", rollbackErr)
			}
		}
	}()

	// Prepare SQL query with upsert logic
	query := `
		INSERT INTO files (
			id, 
			file_metadata, 
			storage_path, 
			processing_status, 
			created_at, 
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET 
			file_metadata = ?,
			storage_path = ?,
			processing_status = ?,
			updated_at = ?
	`

	// Set default values if not provided
	if storage.CreatedAt.IsZero() {
		storage.CreatedAt = time.Now().UTC()
	}
	storage.UpdatedAt = time.Now().UTC()

	// If no processing status is set, default to pending
	if storage.ProcessingStatus == "" {
		storage.ProcessingStatus = "PENDING"
	}

	// Execute query
	_, err = tx.ExecContext(ctx, query,
		// Insert values
		storage.ID,
		storage.FileMetadata,
		storage.StoragePath,
		storage.ProcessingStatus,
		storage.CreatedAt,
		storage.UpdatedAt,
		// Update values
		storage.FileMetadata,
		storage.StoragePath,
		storage.ProcessingStatus,
		storage.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store file metadata: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("File metadata stored successfully: %s", storage.ID)
	return nil
}

// FindByID retrieves file metadata by ID with enhanced security
func (r *SQLiteStorageRepository) FindByID(ctx context.Context, fileID string) (*models.Storage, error) {
	// Validate input
	if fileID == "" {
		return nil, ErrInvalidInput
	}

	// Prepare SQL query
	query := `
		SELECT 
			id, 
			file_metadata, 
			storage_path, 
			processing_status, 
			created_at, 
			updated_at
		FROM files 
		WHERE id = ?
	`

	// Execute query
	row := r.db.QueryRowContext(ctx, query, fileID)

	// Scan result into storage model
	storage := &models.Storage{}
	err := row.Scan(
		&storage.ID,
		&storage.FileMetadata,
		&storage.StoragePath,
		&storage.ProcessingStatus,
		&storage.CreatedAt,
		&storage.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		log.Printf("File not found: %s", fileID)
		return nil, ErrFileNotFound
	} else if err != nil {
		log.Printf("Error retrieving file metadata: %v", err)
		return nil, fmt.Errorf("failed to retrieve file metadata: %v", err)
	}

	return storage, nil
}

// List retrieves files with advanced pagination and filtering
func (r *SQLiteStorageRepository) List(ctx context.Context, userID string, page, pageSize int) ([]*models.Storage, int, error) {
	// Validate input
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Count total files
	countQuery := `SELECT COUNT(*) FROM files WHERE file_metadata->>'$.user_id' = ?`
	var totalFiles int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&totalFiles)
	if err != nil {
		log.Printf("Error counting files: %v", err)
		return nil, 0, fmt.Errorf("failed to count files: %v", err)
	}

	// Retrieve paginated files
	query := `
		SELECT 
			id, 
			file_metadata, 
			storage_path, 
			processing_status, 
			created_at, 
			updated_at
		FROM files 
		WHERE file_metadata->>'$.user_id' = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		log.Printf("Error listing files: %v", err)
		return nil, 0, fmt.Errorf("failed to list files: %v", err)
	}
	defer rows.Close()

	// Process results
	var files []*models.Storage
	for rows.Next() {
		storage := &models.Storage{}
		err := rows.Scan(
			&storage.ID,
			&storage.FileMetadata,
			&storage.StoragePath,
			&storage.ProcessingStatus,
			&storage.CreatedAt,
			&storage.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning file row: %v", err)
			continue
		}
		files = append(files, storage)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error in file rows: %v", err)
		return nil, 0, fmt.Errorf("error processing file rows: %v", err)
	}

	return files, totalFiles, nil
}

// Delete removes a file by ID with user authorization
func (r *SQLiteStorageRepository) Delete(ctx context.Context, fileID, userID string) error {
	// Validate input
	if fileID == "" || userID == "" {
		return ErrInvalidInput
	}

	// Prepare delete query with user authorization
	query := `DELETE FROM files WHERE id = ? AND file_metadata->>'$.user_id' = ?`
	result, err := r.db.ExecContext(ctx, query, fileID, userID)
	if err != nil {
		log.Printf("Error deleting file: %v", err)
		return fmt.Errorf("failed to delete file: %v", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error checking rows affected: %v", err)
		return fmt.Errorf("error checking deletion: %v", err)
	}

	if rowsAffected == 0 {
		log.Printf("No file found with ID %s for user %s", fileID, userID)
		return ErrFileNotFound
	}

	log.Printf("File deleted successfully: %s", fileID)
	return nil
}
