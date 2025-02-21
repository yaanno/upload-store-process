package storage

import (
	"context"
	"io"
)

// Provider defines the interface for file storage operations
type Provider interface {
	// Store saves a file and returns its storage path
	Store(ctx context.Context, fileID string, content io.Reader) (string, error)

	// Retrieve gets a file by its ID
	Retrieve(ctx context.Context, fileID string) (io.ReadCloser, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, fileID string) error
}
