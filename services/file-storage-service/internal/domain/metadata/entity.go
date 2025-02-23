package metadata

import (
	"time"

	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
)

// FileMetadataRecord represents the storage record for file metadata
type FileMetadataRecord struct {
	ID               string
	Metadata         *sharedv1.FileMetadata
	StoragePath      string
	ProcessingStatus string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// FileMetadataListOptions provides filtering and pagination for file metadata listing
type FileMetadataListOptions struct {
	UserID string
	FileID string
	Status string
}

// NewFileMetadataListOptions creates a new FileMetadataListOptions instance
func NewFileMetadataListOptions(userID string) *FileMetadataListOptions {
	return &FileMetadataListOptions{
		UserID: userID,
	}
}

// NewFileMetadataListOptionsWithPagination creates a new FileMetadataListOptions instance
func NewFileMetadataListOptionsWithPagination(userID string, limit int, offset int, sortBy string, sortOrder string) *FileMetadataListOptions {
	return &FileMetadataListOptions{
		UserID: userID,
	}
}
