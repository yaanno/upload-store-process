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

// Migrate performs database schema migrations
func (m *DatabaseMigrator) Migrate(ctx context.Context) error {
	// Enable foreign key support
	_, err := m.db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	if err != nil {
		return fmt.Errorf("failed to enable foreign key support: %v", err)
	}

	// Create files table with JSON metadata support
	createFilesTableQuery := `
	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		file_metadata JSON NOT NULL,
		storage_path TEXT,
		processing_status TEXT DEFAULT 'PENDING',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		UNIQUE(id)
	)`

	_, err = m.db.ExecContext(ctx, createFilesTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create files table: %v", err)
	}

	// Create index for faster queries
	createIndexQuery := `
	CREATE INDEX IF NOT EXISTS idx_files_user_id 
	ON files ((json_extract(file_metadata, '$.user_id')))`

	_, err = m.db.ExecContext(ctx, createIndexQuery)
	if err != nil {
		return fmt.Errorf("failed to create user_id index: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
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
