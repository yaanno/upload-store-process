package configs

import (
	"fmt"
	"os"
	"path/filepath"
)

// DatabaseConfig holds configuration for database connections
type DatabaseConfig struct {
	Driver       string
	DataSourceName string
}

// GetSQLiteConfig generates SQLite database configuration
func GetSQLiteConfig() DatabaseConfig {
	// Determine base path for database storage
	baseDir := os.Getenv("APP_DATA_DIR")
	if baseDir == "" {
		baseDir = filepath.Join(os.Getenv("HOME"), ".upload-store-process")
	}

	// Ensure directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create data directory: %v", err))
	}

	// Construct database path
	dbPath := filepath.Join(baseDir, "file_storage.db")

	return DatabaseConfig{
		Driver: "sqlite3",
		DataSourceName: fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL", dbPath),
	}
}

// GetTestDatabaseConfig generates an in-memory test database configuration
func GetTestDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver: "sqlite3",
		DataSourceName: "file::memory:?cache=shared",
	}
}
