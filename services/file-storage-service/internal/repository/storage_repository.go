package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
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
	ListFiles(ctx context.Context, opts *FileMetadataListOptions) ([]*models.FileMetadataRecord, int, error)
	RemoveFileMetadata(ctx context.Context, fileID string) error
	UpdateFileMetadata(ctx context.Context, metadata *models.FileMetadataRecord) error
}

// FileMetadataListOptions provides filtering and pagination for file metadata listing
type FileMetadataListOptions struct {
	UserID    string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// SQLiteFileMetadataRepository implements FileMetadataRepository for SQLite
type SQLiteFileMetadataRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewSQLiteFileMetadataRepository creates a new SQLite-based file metadata repository
func NewSQLiteFileMetadataRepository(db *sql.DB, logger logger.Logger) *SQLiteFileMetadataRepository {
	return &SQLiteFileMetadataRepository{
		db:     db,
		logger: logger,
	}
}

// UpdateFileMetadata updates an existing file metadata record
func (r *SQLiteFileMetadataRepository) UpdateFileMetadata(ctx context.Context, metadata *models.FileMetadataRecord) error {
	// Validate input
	if metadata == nil {
		return fmt.Errorf("%w: file metadata cannot be nil", ErrInvalidInput)
	}

	if metadata.ID == "" {
		return fmt.Errorf("%w: file ID is required", ErrInvalidInput)
	}

	// Convert FileMetadata to JSON
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
				r.logger.Error().
					Err(rollbackErr).
					Str("original_error", err.Error()).
					Msg("Failed to rollback transaction")
			}
		} else {
			if err = tx.Commit(); err != nil {
				r.logger.Error().
					Err(err).
					Msg("Failed to commit transaction")
			}
		}
	}()

	// Update file metadata
	query := `
		UPDATE file_metadata 
		SET 
			metadata = ?, 
			storage_path = ?, 
			processing_status = ?, 
			updated_at = ?
		WHERE id = ?
	`

	_, err = tx.ExecContext(ctx, query,
		fileMetadataJSON,
		metadata.StoragePath,
		metadata.ProcessingStatus,
		metadata.UpdatedAt,
		metadata.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update file metadata: %w", err)
	}

	return nil
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
				r.logger.Error().
					Err(rollbackErr).
					Str("original_error", err.Error()).
					Msg("Failed to rollback transaction")
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
		r.logger.Error().
			Err(err).
			Str("fileId", metadata.ID).
			Msg("Failed to prepare create file metadata statement")
		return fmt.Errorf("prepare create file metadata: %w", err)
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
		r.logger.Error().
			Err(err).
			Str("fileId", metadata.ID).
			Str("filename", metadata.Metadata.OriginalFilename).
			Msg("Failed to create file metadata")
		return fmt.Errorf("create file metadata: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		r.logger.Error().
			Err(err).
			Str("fileId", metadata.ID).
			Msg("Failed to commit transaction")
		return fmt.Errorf("commit transaction: %w", err)
	}

	r.logger.Info().
		Str("fileId", metadata.ID).
		Str("filename", metadata.Metadata.OriginalFilename).
		Msg("File metadata created successfully")

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
		r.logger.Info().
			Str("fileId", fileID).
			Msg("File metadata not found")
		return nil, ErrFileNotFound
	} else if err != nil {
		r.logger.Error().
			Err(err).
			Str("fileId", fileID).
			Msg("Error retrieving file metadata")
		return nil, fmt.Errorf("failed to retrieve file metadata: %w", err)
	}

	// Unmarshal file metadata
	if len(fileMetadataJSON) > 0 {
		metadata.Metadata = &sharedv1.FileMetadata{}
		if err := json.Unmarshal(fileMetadataJSON, metadata.Metadata); err != nil {
			r.logger.Error().
				Err(err).
				Str("fileId", fileID).
				Msg("Error unmarshaling file metadata")
			return nil, fmt.Errorf("failed to unmarshal file metadata: %w", err)
		}
		metadata.Metadata.UserId = userID
	}

	r.logger.Info().
		Str("fileId", fileID).
		Msg("File metadata retrieved successfully")

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
		r.logger.Error().
			Err(err).
			Str("userId", opts.UserID).
			Msg("Error counting file metadata")
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
		r.logger.Error().
			Err(err).
			Str("userId", opts.UserID).
			Msg("Error listing file metadata")
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
			r.logger.Error().
				Err(err).
				Msg("Error scanning file metadata row")
			continue
		}
		if len(fileMetadataJSON) > 0 {
			metadata.Metadata = &sharedv1.FileMetadata{}
			if err := json.Unmarshal(fileMetadataJSON, metadata.Metadata); err != nil {
				r.logger.Error().
					Err(err).
					Msg("Error unmarshaling file metadata")
				continue
			}
			metadata.Metadata.UserId = userID
		}
		fileMetadataRecords = append(fileMetadataRecords, metadata)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error().
			Err(err).
			Msg("Error in file metadata rows")
		return nil, fmt.Errorf("error processing file metadata rows: %w", err)
	}

	r.logger.Info().
		Str("userId", opts.UserID).
		Int("totalFiles", totalFiles).
		Msg("File metadata listed successfully")

	return fileMetadataRecords, nil
}

// ListFiles retrieves file metadata with advanced pagination and filtering
func (r *SQLiteFileMetadataRepository) ListFiles(ctx context.Context, opts *FileMetadataListOptions) ([]*models.FileMetadataRecord, int, error) {
	// Validate input
	if opts == nil {
		return nil, 0, fmt.Errorf("%w: list options cannot be nil", ErrInvalidInput)
	}

	if opts.UserID == "" {
		return nil, 0, fmt.Errorf("%w: user ID is required", ErrInvalidInput)
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
		r.logger.Error().
			Err(err).
			Str("userId", opts.UserID).
			Msg("Error counting file metadata")
		return nil, 0, fmt.Errorf("failed to count file metadata: %w", err)
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
		r.logger.Error().
			Err(err).
			Str("userId", opts.UserID).
			Msg("Error listing file metadata")
		return nil, 0, fmt.Errorf("failed to list file metadata: %w", err)
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
			r.logger.Error().
				Err(err).
				Msg("Error scanning file metadata row")
			continue
		}
		if len(fileMetadataJSON) > 0 {
			metadata.Metadata = &sharedv1.FileMetadata{}
			if err := json.Unmarshal(fileMetadataJSON, metadata.Metadata); err != nil {
				r.logger.Error().
					Err(err).
					Msg("Error unmarshaling file metadata")
				continue
			}
			metadata.Metadata.UserId = userID
		}
		fileMetadataRecords = append(fileMetadataRecords, metadata)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error().
			Err(err).
			Msg("Error in file metadata rows")
		return nil, 0, fmt.Errorf("error processing file metadata rows: %w", err)
	}

	r.logger.Info().
		Str("userId", opts.UserID).
		Int("totalFiles", totalFiles).
		Msg("File metadata listed successfully")

	return fileMetadataRecords, totalFiles, nil
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
		r.logger.Error().
			Err(err).
			Str("fileId", fileID).
			Msg("Error removing file metadata")
		return fmt.Errorf("failed to remove file metadata: %w", ErrDatabaseOperation)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("fileId", fileID).
			Msg("Error checking rows affected")
		return fmt.Errorf("error checking deletion: %w", ErrDatabaseOperation)
	}

	if rowsAffected == 0 {
		r.logger.Info().
			Str("fileId", fileID).
			Msg("No file metadata found with ID")
		return ErrFileNotFound
	}

	r.logger.Info().
		Str("fileId", fileID).
		Msg("File metadata removed successfully")

	return nil
}

var _ FileMetadataRepository = (*SQLiteFileMetadataRepository)(nil)
