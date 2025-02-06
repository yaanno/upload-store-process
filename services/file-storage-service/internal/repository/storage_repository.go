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

// FileMetadataRepository defines the interface for file metadata storage operations
type FileMetadataRepository interface {
	CreateFileMetadata(ctx context.Context, metadata *models.FileMetadataRecord) error
	RetrieveFileMetadataByID(ctx context.Context, fileID string) (*models.FileMetadataRecord, error)
	ListFileMetadata(ctx context.Context, opts *FileMetadataListOptions) ([]*models.FileMetadataRecord, error)
	RemoveFileMetadata(ctx context.Context, fileID string) error
}

// FileMetadataListOptions provides filtering and pagination for file metadata listing
type FileMetadataListOptions struct {
	UserID     string
	Limit      int
	Offset     int
	SortBy     string
	SortOrder  string
}

// SQLiteFileMetadataRepository implements FileMetadataRepository for SQLite
type SQLiteFileMetadataRepository struct {
	db       *sql.DB
	migrator *DatabaseMigrator
	logger   *slog.Logger
}

// NewSQLiteFileMetadataRepository creates a new SQLite-based file metadata repository
func NewSQLiteFileMetadataRepository(migrator *DatabaseMigrator, logger *slog.Logger) *SQLiteFileMetadataRepository {
	return &SQLiteFileMetadataRepository{
		db:       migrator.GetDB(),
		migrator: migrator,
		logger:   logger,
	}
}

// CreateFileMetadata saves file metadata with transaction support and upsert logic
func (r *SQLiteFileMetadataRepository) CreateFileMetadata(ctx context.Context, metadata *models.FileMetadataRecord) error {
	// Validate input with more robust checks
	if metadata == nil {
		return fmt.Errorf("%w: file metadata cannot be nil", ErrInvalidInput)
	}

	if metadata.ID == "" {
		return fmt.Errorf("%w: file ID is required", ErrInvalidInput)
	}

	// Validate metadata model
	if err := metadata.Validate(); err != nil {
		return fmt.Errorf("invalid file metadata: %w", err)
	}

	// Convert FileMetadata to JSON with error handling
	var fileMetadataJSON []byte
	var err error
	if metadata.Metadata != nil {
		fileMetadataJSON, err = json.Marshal(metadata.Metadata)
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
		INSERT INTO file_metadata (
			id, 
			metadata_json, 
			storage_path, 
			processing_status, 
			user_id,
			created_at, 
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET 
			metadata_json = ?,
			storage_path = ?,
			processing_status = ?,
			updated_at = ?
	`

	// Set default values if not provided
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = time.Now().UTC()
	}
	metadata.UpdatedAt = time.Now().UTC()

	// If no processing status is set, default to pending
	if metadata.ProcessingStatus == "" {
		metadata.ProcessingStatus = "PENDING"
	}

	// Ensure user_id is extracted
	userID := ""
	if metadata.Metadata != nil {
		userID = metadata.Metadata.UserId
	}

	// Execute query with prepared statement for better performance
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", ErrDatabaseOperation)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		// Insert values
		metadata.ID,
		fileMetadataJSON,
		metadata.StoragePath,
		metadata.ProcessingStatus,
		userID,
		metadata.CreatedAt,
		metadata.UpdatedAt,
		// Update values
		fileMetadataJSON,
		metadata.StoragePath,
		metadata.ProcessingStatus,
		metadata.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store file metadata: %w", ErrDatabaseOperation)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", ErrDatabaseOperation)
	}

	r.logger.Info("File metadata stored successfully", 
		slog.String("file_id", metadata.ID),
		slog.String("storage_path", metadata.StoragePath))
	return nil
}

// RetrieveFileMetadataByID retrieves file metadata by ID with enhanced security
func (r *SQLiteFileMetadataRepository) RetrieveFileMetadataByID(ctx context.Context, fileID string) (*models.FileMetadataRecord, error) {
	// Validate input
	if fileID == "" {
		return nil, fmt.Errorf("%w: file ID cannot be empty", ErrInvalidInput)
	}

	// Prepare SQL query
	query := `
		SELECT 
			id, 
			metadata_json, 
			storage_path, 
			processing_status, 
			user_id,
			created_at, 
			updated_at
		FROM file_metadata 
		WHERE id = ?
	`

	// Execute query
	row := r.db.QueryRowContext(ctx, query, fileID)

	// Scan result into metadata model
	metadata := &models.FileMetadataRecord{}
	var fileMetadataJSON []byte
	var userID string
	err := row.Scan(
		&metadata.ID,
		&fileMetadataJSON,
		&metadata.StoragePath,
		&metadata.ProcessingStatus,
		&userID,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		r.logger.Info("File metadata not found", 
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
		metadata.Metadata = &sharedv1.FileMetadata{}
		if err := json.Unmarshal(fileMetadataJSON, metadata.Metadata); err != nil {
			r.logger.Error("Error unmarshaling file metadata", 
				slog.String("file_id", fileID),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to unmarshal file metadata: %w", err)
		}
		metadata.Metadata.UserId = userID
	}

	r.logger.Info("File metadata retrieved successfully", 
		slog.String("file_id", fileID))
	return metadata, nil
}

// ListFileMetadata retrieves file metadata with advanced pagination and filtering
func (r *SQLiteFileMetadataRepository) ListFileMetadata(ctx context.Context, opts *FileMetadataListOptions) ([]*models.FileMetadataRecord, error) {
	// Validate input
	if opts == nil {
		return nil, fmt.Errorf("%w: list options cannot be nil", ErrInvalidInput)
	}

	if opts.UserID == "" {
		return nil, fmt.Errorf("%w: user ID is required", ErrInvalidInput)
	}

	// Normalize pagination
	if opts.Limit < 1 || opts.Limit > 100 {
		opts.Limit = 10
	}

	// Calculate offset
	offset := (opts.Offset - 1) * opts.Limit

	// Count total files
	countQuery := `SELECT COUNT(*) FROM file_metadata WHERE user_id = ?`
	var totalFiles int
	err := r.db.QueryRowContext(ctx, countQuery, opts.UserID).Scan(&totalFiles)
	if err != nil {
		r.logger.Error("Error counting file metadata", 
			slog.String("user_id", opts.UserID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to count file metadata: %w", err)
	}

	// Retrieve paginated file metadata
	query := `
		SELECT 
			id, 
			metadata_json, 
			storage_path, 
			processing_status, 
			user_id,
			created_at, 
			updated_at
		FROM file_metadata
		WHERE user_id = ?
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, opts.UserID, opts.Limit, offset)
	if err != nil {
		r.logger.Error("Error listing file metadata", 
			slog.String("user_id", opts.UserID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to list file metadata: %w", err)
	}
	defer rows.Close()

	var fileMetadataRecords []*models.FileMetadataRecord
	for rows.Next() {
		metadata := &models.FileMetadataRecord{}
		var fileMetadataJSON []byte
		var userID string
		err := rows.Scan(
			&metadata.ID,
			&fileMetadataJSON,
			&metadata.StoragePath,
			&metadata.ProcessingStatus,
			&userID,
			&metadata.CreatedAt,
			&metadata.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Error scanning file metadata row", 
				slog.String("error", err.Error()))
			continue
		}
		if len(fileMetadataJSON) > 0 {
			metadata.Metadata = &sharedv1.FileMetadata{}
			if err := json.Unmarshal(fileMetadataJSON, metadata.Metadata); err != nil {
				r.logger.Error("Error unmarshaling file metadata", 
					slog.String("error", err.Error()))
				continue
			}
			metadata.Metadata.UserId = userID
		}
		fileMetadataRecords = append(fileMetadataRecords, metadata)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error in file metadata rows", 
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("error processing file metadata rows: %w", err)
	}

	r.logger.Info("File metadata listed successfully", 
		slog.String("user_id", opts.UserID),
		slog.Int("total_files", totalFiles))
	return fileMetadataRecords, nil
}

// RemoveFileMetadata removes a file metadata record by ID
func (r *SQLiteFileMetadataRepository) RemoveFileMetadata(ctx context.Context, fileID string) error {
	// Validate input
	if fileID == "" {
		return fmt.Errorf("%w: file ID cannot be empty", ErrInvalidInput)
	}

	// Prepare delete query
	query := `DELETE FROM file_metadata WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, fileID)
	if err != nil {
		r.logger.Error("Error removing file metadata", 
			slog.String("file_id", fileID),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to remove file metadata: %w", ErrDatabaseOperation)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Error checking rows affected", 
			slog.String("file_id", fileID),
			slog.String("error", err.Error()))
		return fmt.Errorf("error checking deletion: %w", ErrDatabaseOperation)
	}

	if rowsAffected == 0 {
		r.logger.Info("No file metadata found with ID", 
			slog.String("file_id", fileID))
		return ErrFileNotFound
	}

	r.logger.Info("File metadata removed successfully", 
		slog.String("file_id", fileID))
	return nil
}
