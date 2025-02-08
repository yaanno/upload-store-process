package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DatabaseMigrator handles SQLite database initialization and migrations
type DatabaseMigrator struct {
	db *sql.DB
}

// NewDatabaseMigrator creates a new database migrator
func NewDatabaseMigrator(dataSourceName string) (*DatabaseMigrator, error) {
	// Open database connection
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Configure database connection
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DatabaseMigrator{db: db}, nil
}

// Migrate performs database migration
func (m *DatabaseMigrator) Migrate(ctx context.Context) error {
	// Enable foreign key support
	_, err := m.db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	if err != nil {
		return fmt.Errorf("failed to enable foreign key support: %v", err)
	}

	// Create file_metadata table
	createFileMetadataTableQuery := `
	CREATE TABLE IF NOT EXISTS file_metadata (
		id TEXT PRIMARY KEY,
		metadata_json TEXT NOT NULL,
		storage_path TEXT NOT NULL,
		processing_status TEXT NOT NULL,
		user_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME,
		is_deleted BOOLEAN DEFAULT 0
	)`

	// Create index for faster user_id queries
	createUserIdIndexQuery := `
	CREATE INDEX IF NOT EXISTS idx_file_metadata_user_id 
	ON file_metadata (user_id)`

	// Create files table with reference to file_metadata
	createFilesTableQuery := `
	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		file_metadata_id TEXT NOT NULL,
		storage_path TEXT,
		processing_status TEXT DEFAULT 'PENDING',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		UNIQUE(id),
		FOREIGN KEY (file_metadata_id) REFERENCES file_metadata (id)
	)`

	// Begin transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}

	// Execute migrations
	migrationQueries := []string{
		createFileMetadataTableQuery,
		createUserIdIndexQuery,
		createFilesTableQuery,
	}

	for _, query := range migrationQueries {
		_, err = tx.ExecContext(ctx, query)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration query: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// CreateTables is now deprecated, kept for backwards compatibility
func (m *DatabaseMigrator) CreateTables() error {
	return m.Migrate(context.Background())
}

// Close closes the database connection
func (m *DatabaseMigrator) Close() error {
	return m.db.Close()
}

// GetDB returns the underlying database connection
func (m *DatabaseMigrator) GetDB() *sql.DB {
	return m.db
}

// InitializeTestDatabase creates a temporary in-memory database for testing
func InitializeTestDatabase() (*DatabaseMigrator, error) {
	// Use in-memory database for testing
	migrator, err := NewDatabaseMigrator("file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}

	// Run migrations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := migrator.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %v", err)
	}

	return migrator, nil
}
