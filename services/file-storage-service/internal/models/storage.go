package models

import (
	"errors"
	"time"

	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
)

// FileMetadataRecord represents the storage record for file metadata
type FileMetadataRecord struct {
	// ID is the unique identifier for the file
	ID string

	// Metadata contains the detailed file metadata
	Metadata *sharedv1.FileMetadata

	// StoragePath is the location where the file is stored
	StoragePath string

	// ProcessingStatus indicates the current status of file processing
	ProcessingStatus string

	// CreatedAt is the timestamp when the record was first created
	CreatedAt time.Time

	// UpdatedAt is the timestamp of the last update to the record
	UpdatedAt time.Time
}

// Validate checks the integrity of the FileMetadataRecord
func (f *FileMetadataRecord) Validate() error {
	if f == nil {
		return ErrNilRecord
	}

	if f.ID == "" {
		return ErrEmptyID
	}

	if f.Metadata == nil {
		return ErrNilMetadata
	}

	return nil
}

// Error variables for validation
var (
	// ErrNilRecord indicates that the record is nil
	ErrNilRecord = errors.New("file metadata record is nil")

	// ErrEmptyID indicates that the record has an empty ID
	ErrEmptyID = errors.New("file metadata record ID cannot be empty")

	// ErrNilMetadata indicates that the metadata is nil
	ErrNilMetadata = errors.New("file metadata cannot be nil")
)
