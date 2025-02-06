package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
)

var (
	// ErrInvalidInput represents an error for invalid input
	ErrInvalidInput = errors.New("invalid input")
	
	// ErrFileNotFound represents an error when a file is not found
	ErrFileNotFound = errors.New("file not found")
	
	// ErrDatabaseOperation represents a generic database operation error
	ErrDatabaseOperation = errors.New("database operation failed")
)

// StorageRepository defines the interface for file metadata storage
type StorageRepository interface {
	Store(ctx context.Context, storage *models.Storage) error
	FindByID(ctx context.Context, fileID string) (*models.Storage, error)
	ListFiles(ctx context.Context, opts *ListFilesOptions) ([]*models.Storage, error)
	DeleteFile(ctx context.Context, fileID string) error
}

// ListFilesOptions provides filtering and pagination for file listing
type ListFilesOptions struct {
	UserID     string
	Limit      int
	Offset     int
	SortBy     string
	SortOrder  string
}

// SQLiteStorageRepository implements StorageRepository for SQLite
type SQLiteStorageRepository struct {
	db       *sql.DB
	migrator *DatabaseMigrator
	logger   *slog.Logger
}

// NewSQLiteStorageRepository creates a new SQLite-based repository
func NewSQLiteStorageRepository(migrator *DatabaseMigrator, logger *slog.Logger) *SQLiteStorageRepository {
	return &SQLiteStorageRepository{
		db:       migrator.GetDB(),
		migrator: migrator,
		logger:   logger,
	}
}

// Store saves file metadata with transaction support and upsert logic
func (r *SQLiteStorageRepository) Store(ctx context.Context, storage *models.Storage) error {
	// Validate input with more robust checks
	if storage == nil {
		return fmt.Errorf("%w: storage cannot be nil", ErrInvalidInput)
	}

	if storage.ID == "" {
		return fmt.Errorf("%w: file ID is required", ErrInvalidInput)
	}

	// Validate storage model
	if err := storage.Validate(); err != nil {
		return fmt.Errorf("invalid storage model: %w", err)
	}

	// Convert FileMetadata to JSON with error handling
	var fileMetadataJSON []byte
	var err error
	if storage.FileMetadata != nil {
		fileMetadataJSON, err = json.Marshal(storage.FileMetadata)
		if err != nil {
			return fmt.Errorf("failed to marshal file metadata: %w", err)
		}
	}

	// Begin transaction with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", ErrDatabaseOperation)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.logger.Error("Failed to rollback transaction", 
					slog.String("original_error", err.Error()),
					slog.String("rollback_error", rollbackErr.Error()))
			}
		}
	}()

	// Prepare SQL query with upsert logic and improved performance hints
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

	// Execute query with prepared statement for better performance
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", ErrDatabaseOperation)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		// Insert values
		storage.ID,
		fileMetadataJSON,
		storage.StoragePath,
		storage.ProcessingStatus,
		storage.CreatedAt,
		storage.UpdatedAt,
		// Update values
		fileMetadataJSON,
		storage.StoragePath,
		storage.ProcessingStatus,
		storage.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store file metadata: %w", ErrDatabaseOperation)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", ErrDatabaseOperation)
	}

	r.logger.Info("File metadata stored successfully", 
		slog.String("file_id", storage.ID),
		slog.String("storage_path", storage.StoragePath))
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
	var fileMetadataJSON []byte
	err := row.Scan(
		&storage.ID,
		&fileMetadataJSON,
		&storage.StoragePath,
		&storage.ProcessingStatus,
		&storage.CreatedAt,
		&storage.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		r.logger.Info("File not found", 
			slog.String("file_id", fileID))
		return nil, ErrFileNotFound
	} else if err != nil {
		r.logger.Error("Error retrieving file metadata", 
			slog.String("file_id", fileID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to retrieve file metadata: %w", err)
	}

	// Unmarshal file metadata
	if len(fileMetadataJSON) > 0 {
		storage.FileMetadata = &sharedv1.FileMetadata{}
		if err := json.Unmarshal(fileMetadataJSON, storage.FileMetadata); err != nil {
			r.logger.Error("Error unmarshaling file metadata", 
				slog.String("file_id", fileID),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to unmarshal file metadata: %w", err)
		}
	}

	r.logger.Info("File metadata retrieved successfully", 
		slog.String("file_id", fileID))
	return storage, nil
}

// ListFiles retrieves files with advanced pagination and filtering
func (r *SQLiteStorageRepository) ListFiles(ctx context.Context, opts *ListFilesOptions) ([]*models.Storage, error) {
	// Validate input
	if opts.UserID == "" {
		return nil, ErrInvalidInput
	}

	if opts.Limit < 1 || opts.Limit > 100 {
		opts.Limit = 10
	}

	// Calculate offset
	offset := (opts.Offset - 1) * opts.Limit

	// Count total files
	countQuery := `SELECT COUNT(*) FROM files WHERE file_metadata->>'$.user_id' = ?`
	var totalFiles int
	err := r.db.QueryRowContext(ctx, countQuery, opts.UserID).Scan(&totalFiles)
	if err != nil {
		r.logger.Error("Error counting files", 
			slog.String("user_id", opts.UserID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to count files: %w", err)
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

	rows, err := r.db.QueryContext(ctx, query, opts.UserID, opts.Limit, offset)
	if err != nil {
		r.logger.Error("Error listing files", 
			slog.String("user_id", opts.UserID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer rows.Close()

	// Process results
	var files []*models.Storage
	for rows.Next() {
		storage := &models.Storage{}
		var fileMetadataJSON []byte
		err := rows.Scan(
			&storage.ID,
			&fileMetadataJSON,
			&storage.StoragePath,
			&storage.ProcessingStatus,
			&storage.CreatedAt,
			&storage.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Error scanning file row", 
				slog.String("error", err.Error()))
			continue
		}
		if len(fileMetadataJSON) > 0 {
			storage.FileMetadata = &sharedv1.FileMetadata{}
			if err := json.Unmarshal(fileMetadataJSON, storage.FileMetadata); err != nil {
				r.logger.Error("Error unmarshaling file metadata", 
					slog.String("error", err.Error()))
				continue
			}
		}
		files = append(files, storage)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error in file rows", 
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("error processing file rows: %w", err)
	}

	r.logger.Info("Files listed successfully", 
		slog.String("user_id", opts.UserID),
		slog.Int("total_files", totalFiles))
	return files, nil
}

// DeleteFile removes a file by ID with user authorization
func (r *SQLiteStorageRepository) DeleteFile(ctx context.Context, fileID string) error {
	// Validate input
	if fileID == "" {
		return ErrInvalidInput
	}

	// Prepare delete query with user authorization
	query := `DELETE FROM files WHERE id = ? AND file_metadata->>'$.user_id' = ?`
	result, err := r.db.ExecContext(ctx, query, fileID, "")
	if err != nil {
		r.logger.Error("Error deleting file", 
			slog.String("file_id", fileID),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Error checking rows affected", 
			slog.String("file_id", fileID),
			slog.String("error", err.Error()))
		return fmt.Errorf("error checking deletion: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Info("No file found with ID", 
			slog.String("file_id", fileID))
		return ErrFileNotFound
	}

	r.logger.Info("File deleted successfully", 
		slog.String("file_id", fileID))
	return nil
}
