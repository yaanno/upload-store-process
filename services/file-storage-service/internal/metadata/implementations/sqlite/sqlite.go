package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

type CleanupResult struct {
	DeletedCount    int64
	LastProcessedID string
	FailedIDs       []string
	Error           error
}

var (
	// ErrInvalidInput represents an error for invalid input
	ErrInvalidInput = errors.New("invalid input")

	// ErrFileNotFound represents an error when a file is not found
	ErrFileNotFound = errors.New("file not found")

	// ErrDatabaseOperation represents a generic database operation error
	ErrDatabaseOperation = errors.New("database operation failed")
)

// SQLiteFileMetadataRepository implements FileMetadataRepository for SQLite
type SQLiteFileMetadataRepository struct {
	db     *sql.DB
	logger *logger.Logger
	mu     sync.RWMutex
	// lockTimeout is the maximum duration to acquire the lock
	lockTimeout time.Duration
}

// NewSQLiteFileMetadataRepository creates a new SQLite-based file metadata repository
func NewSQLiteFileMetadataRepository(db *sql.DB, logger *logger.Logger) *SQLiteFileMetadataRepository {
	return &SQLiteFileMetadataRepository{
		db:          db,
		logger:      logger,
		lockTimeout: 5 * time.Second,
	}
}

func (r *SQLiteFileMetadataRepository) BeginTx(ctx context.Context) (interface{}, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

func (r *SQLiteFileMetadataRepository) CommitTx(ctx context.Context, tx interface{}) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}
	return sqlTx.Commit()
}

func (r *SQLiteFileMetadataRepository) RollbackTx(ctx context.Context, tx interface{}) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}
	return sqlTx.Rollback()
}

// UpdateFileMetadata updates an existing file metadata record
func (r *SQLiteFileMetadataRepository) UpdateFileMetadata(ctx context.Context, metadata *domain.FileMetadataRecord) error {

	// Validate metadata model
	if err := metadata.Validate(); err != nil {
		return fmt.Errorf("invalid file metadata: %w", err)
	}

	if err := r.acquireLock(ctx); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.mu.Unlock()

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
			metadata_json = ?, 
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
func (r *SQLiteFileMetadataRepository) CreateFileMetadata(ctx context.Context, metadata *domain.FileMetadataRecord) error {

	// Validate metadata model
	if err := metadata.Validate(); err != nil {
		return fmt.Errorf("invalid file metadata: %w", err)
	}

	if err := r.acquireLock(ctx); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.mu.Unlock()

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
func (r *SQLiteFileMetadataRepository) RetrieveFileMetadataByID(ctx context.Context, fileID string) (*domain.FileMetadataRecord, error) {
	// Validate input
	if fileID == "" {
		return nil, fmt.Errorf("%w: file ID cannot be empty", ErrInvalidInput)
	}

	if err := r.acquireLock(ctx); err != nil {
		return &domain.FileMetadataRecord{}, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.mu.Unlock()

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
	metadata := &domain.FileMetadataRecord{}
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

// ListFileMetadata retrieves file metadata based on provided options
func (r *SQLiteFileMetadataRepository) ListFileMetadata(ctx context.Context, opts *domain.FileMetadataListOptions) ([]*domain.FileMetadataRecord, error) {
	if err := opts.ValidateEssential(); err != nil {
		return nil, fmt.Errorf("invalid list options: %w", err)
	}

	if err := r.acquireLock(ctx); err != nil {
		return []*domain.FileMetadataRecord{}, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.mu.Unlock()

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
	`

	rows, err := r.db.QueryContext(ctx, query, opts.UserID)
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("userId", opts.UserID).
			Msg("Error listing file metadata")
		return nil, fmt.Errorf("failed to list file metadata: %w", err)
	}
	defer rows.Close()

	var fileMetadataRecords []*domain.FileMetadataRecord
	for rows.Next() {
		metadata := &domain.FileMetadataRecord{}
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

// ListFiles retrieves file metadata based on provided options
func (r *SQLiteFileMetadataRepository) ListFiles(ctx context.Context, opts *domain.FileMetadataListOptions) ([]*domain.FileMetadataRecord, int, error) {

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, 0, fmt.Errorf("invalid list options: %w", err)
	}

	if err := r.acquireLock(ctx); err != nil {
		return []*domain.FileMetadataRecord{}, 0, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.mu.Unlock()

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Count total files
	countQuery := `SELECT COUNT(*) FROM file_metadata WHERE user_id = ?`
	var totalFiles int

	err = tx.QueryRowContext(ctx, countQuery, opts.UserID).Scan(&totalFiles)
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("userId", opts.UserID).
			Msg("Error counting file metadata")
		return nil, 0, fmt.Errorf("failed to count file metadata: %w", err)
	}

	// Retrieve file metadata
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
		AND (? = '' OR processing_status = ?)
        AND (is_deleted = 0 OR is_deleted IS NULL)
        ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := tx.QueryContext(ctx, query, opts.UserID, opts.Status, opts.Status, 10)
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("userId", opts.UserID).
			Msg("Error listing file metadata")
		return nil, 0, fmt.Errorf("failed to list file metadata: %w", err)
	}
	defer rows.Close()

	var fileMetadataRecords []*domain.FileMetadataRecord
	for rows.Next() {
		metadata := &domain.FileMetadataRecord{}
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

	return fileMetadataRecords, totalFiles, tx.Commit()
}

// RemoveFileMetadata removes a file metadata record by ID
func (r *SQLiteFileMetadataRepository) RemoveFileMetadata(ctx context.Context, fileID string) error {
	// Validate input
	if fileID == "" {
		return fmt.Errorf("%w: file ID cannot be empty", ErrInvalidInput)
	}

	if err := r.acquireLock(ctx); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.mu.Unlock()

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

// IsFileOwnedByUser checks if a file is owned by a user
func (r *SQLiteFileMetadataRepository) IsFileOwnedByUser(ctx context.Context, opts *domain.FileMetadataListOptions) (bool, error) {
	// Validate input
	if err := opts.ValidateEssential(); err != nil {
		return false, fmt.Errorf("invalid list options: %w", err)
	}

	// Select query: count number of files with given ID and user ID

	query := `SELECT COUNT(*) FROM file_metadata WHERE id = ? AND user_id = ?`
	var count int
	row := r.db.QueryRowContext(ctx, query, opts.FileID, opts.UserID)
	if row.Err() != nil {
		r.logger.Error().
			Err(row.Err()).
			Str("fileId", opts.FileID).
			Str("userId", opts.UserID).
			Msg("Error checking file ownership")
		return false, fmt.Errorf("failed to check file ownership: %w", ErrDatabaseOperation)
	}

	if err := row.Scan(&count); err != nil {
		r.logger.Error().
			Err(err).
			Str("fileId", opts.FileID).
			Str("userId", opts.UserID).
			Msg("Error scanning file ownership count")
		return false, fmt.Errorf("failed to check file ownership: %w", ErrDatabaseOperation)
	}
	return count > 0, nil
}

// SoftDeleteMetadata marks a file metadata record as deleted
func (r *SQLiteFileMetadataRepository) SoftDeleteMetadata(ctx context.Context, fileID, userID string) error {
	if err := r.acquireLock(ctx); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.mu.Unlock()

	query := `
        UPDATE file_metadata 
        SET 
            deleted_at = CURRENT_TIMESTAMP, 
            is_deleted = 1 
        WHERE id = ? AND user_id = ?
    `
	_, err := r.db.ExecContext(ctx, query, fileID, userID)
	return err
}

func (r *SQLiteFileMetadataRepository) CleanupExpiredMetadata(ctx context.Context, expiredBefore time.Time) (int64, error) {
	result := &CleanupResult{}
	batchSize := 100
	var lastID string // Cursor for pagination
	if err := r.acquireLock(ctx); err != nil {
		return 0, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.mu.Unlock()

	for {
		// Start transaction for this batch
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Defer transaction handling
		defer func() {
			if p := recover(); p != nil {
				_ = tx.Rollback()
				panic(p)
			} else if err != nil {
				_ = tx.Rollback()
			}
		}()

		query := `
            DELETE FROM file_metadata 
            WHERE id IN (
                SELECT id 
                FROM file_metadata 
                WHERE processing_status = 'PENDING' 
                AND created_at < ? 
                AND updated_at < ?
                AND id > ?
                ORDER BY id
                LIMIT ?
            )
            RETURNING id
        `

		// Collect deleted IDs to track progress
		rows, err := tx.QueryContext(ctx, query, expiredBefore, expiredBefore, lastID, batchSize)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to delete expired metadata: %w", err)
		}

		var deletedIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				tx.Rollback()
				result.FailedIDs = append(result.FailedIDs, id)
				continue
			}
			deletedIDs = append(deletedIDs, id)
		}
		rows.Close()

		if err = tx.Commit(); err != nil {
			return 0, fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Update progress
		if len(deletedIDs) > 0 {
			result.DeletedCount += int64(len(deletedIDs))
			lastID = deletedIDs[len(deletedIDs)-1]
			result.LastProcessedID = lastID
		}

		// Break if we processed less than batch size
		if len(deletedIDs) < batchSize {
			break
		}

		r.logger.Info().
			Int("batchSize", len(deletedIDs)).
			Interface("deleted", result).
			Str("lastID", result.LastProcessedID).
			Int64("totalDeleted", result.DeletedCount).
			Msg("Batch cleanup completed")
	}

	return result.DeletedCount, nil
}

// FindExpiredUploads finds metadata records with expired upload tokens
func (r *SQLiteFileMetadataRepository) FindExpiredUploads(ctx context.Context, expiredBefore time.Time) ([]*domain.FileMetadataRecord, int64, error) {
	query := `
        SELECT id, metadata_json, storage_path, processing_status, user_id, created_at, updated_at
        FROM file_metadata 
        WHERE processing_status = 'PENDING' 
        AND created_at < ?
        AND updated_at < ?
    `
	rows, err := r.db.QueryContext(ctx, query, expiredBefore, expiredBefore)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query expired uploads: %w", err)
	}
	defer rows.Close()

	var fileMetadataRecords []*domain.FileMetadataRecord
	for rows.Next() {
		metadata := &domain.FileMetadataRecord{}
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

	totalFiles := int64(len(fileMetadataRecords))

	r.logger.Info().
		Int64("totalFiles", totalFiles).
		Msg("File metadata listed successfully")

	return fileMetadataRecords, totalFiles, nil
}

func (r *SQLiteFileMetadataRepository) acquireLock(ctx context.Context) error {
	lockChan := make(chan struct{})

	go func() {
		r.mu.Lock()
		close(lockChan)
	}()

	select {
	case <-lockChan:
		return nil
	case <-time.After(r.lockTimeout):
		return fmt.Errorf("lock acquisition timeout: possible deadlock detected")
	case <-ctx.Done():
		return ctx.Err()
	}
}
