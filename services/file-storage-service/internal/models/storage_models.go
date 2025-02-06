package models

import (
	"fmt"
	"time"

	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
)

// Storage represents the internal model for file metadata
type Storage struct {
	ID               string
	FileMetadata     *sharedv1.FileMetadata
	StoragePath      string
	ProcessingStatus string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validate performs basic validation on the storage model
func (s *Storage) Validate() error {
	if s.FileMetadata == nil {
		return fmt.Errorf("file metadata is required")
	}

	if s.FileMetadata.OriginalFilename == "" {
		return fmt.Errorf("original filename is required")
	}

	return nil
}
