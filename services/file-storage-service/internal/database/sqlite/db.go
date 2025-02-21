package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// NewDatabaseMigrator creates a new database migrator
func NewDatabase(ctx context.Context, dataSourceName string) (*sql.DB, error) {
	// Open database connection
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Configure database connection
	// update with pooling later
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping the database to ensure a connection is established
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	migrator, err := NewDatabaseMigrator(db)
	if err != nil {
		return nil, err
	}
	// Run migrations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := migrator.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return db, nil
}

// InitializeTestDatabase creates a temporary in-memory database for testing
func InitializeTestDatabase(ctx context.Context) (*sql.DB, error) {
	// Use in-memory database for testing
	db, err := NewDatabase(ctx, "file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize test database: %v", err)
	}
	migrator, err := NewDatabaseMigrator(db)
	if err != nil {
		return nil, err
	}

	// Run migrations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := migrator.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %v", err)
	}

	return db, nil
}
