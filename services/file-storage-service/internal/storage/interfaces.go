package storage

import (
	"context"
	"io"
)

// FileStorageProvider defines the contract for file storage mechanisms
type FileStorageProvider interface {
	// StoreFile stores a file and returns its relative storage path
	StoreFile(
		ctx context.Context, 
		fileID string, 
		originalFilename string, 
		fileReader io.Reader,
	) (storagePath string, err error)

	// RetrieveFile retrieves a file by its storage path
	RetrieveFile(
		ctx context.Context, 
		storagePath string,
	) (fileReader io.Reader, err error)

	// DeleteFile removes a file from storage
	DeleteFile(
		ctx context.Context, 
		storagePath string,
	) error

	// GenerateStoragePath creates a consistent path for file storage
	GenerateStoragePath(
		fileID string, 
		originalFilename string,
	) string
}
