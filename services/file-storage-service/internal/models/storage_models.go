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

	if s.FileMetadata.FileSizeBytes <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}

	if s.FileMetadata.FileType == "" {
		return fmt.Errorf("file type is required")
	}

	if s.FileMetadata.UploadTimestamp <= 0 {
		return fmt.Errorf("upload timestamp must be greater than 0")
	}

	if s.FileMetadata.StoragePath == "" {
		return fmt.Errorf("storage path is required")
	}

	if s.FileMetadata.UserId == "" {
		return fmt.Errorf("user ID is required")
	}

	if s.ProcessingStatus == "" {
		return fmt.Errorf("processing status is required")
	}

	if s.StoragePath == "" {
		return fmt.Errorf("storage path is required")
	}

	return nil
}

func (s *Storage) String() string {
	return fmt.Sprintf("ID: %s, FileMetadata: %s, StoragePath: %s, ProcessingStatus: %s, CreatedAt: %s, UpdatedAt: %s",
		s.ID, s.FileMetadata, s.StoragePath, s.ProcessingStatus, s.CreatedAt, s.UpdatedAt)
}
