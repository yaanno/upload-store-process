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

// FileMetadataListOptions provides filtering and pagination for file metadata listing
type FileMetadataListOptions struct {
	UserID    string
	FileID    string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// NewFileMetadataListOptions creates a new FileMetadataListOptions instance with default values
func NewFileMetadataListOptions(userID string) *FileMetadataListOptions {
	return &FileMetadataListOptions{
		UserID:    userID,
		Limit:     10,
		Offset:    1,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// NewFileMetadataListOptionsWithPagination creates a new FileMetadataListOptions instance with default values
func NewFileMetadataListOptionsWithPagination(userID string, limit int, offset int, sortBy string, sortOrder string) *FileMetadataListOptions {
	return &FileMetadataListOptions{
		UserID:    userID,
		Limit:     limit,
		Offset:    offset,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}
}

func (o *FileMetadataListOptions) Default() {
	if o.Limit == 0 {
		o.Limit = 10
	}

	if o.Offset == 0 {
		o.Offset = 1
	}
}

// Validate checks the integrity of the FileMetadataListOptions
func (o *FileMetadataListOptions) Validate() error {
	if err := o.ValidateEssential(); err != nil {
		return err
	}

	if o.SortBy == "" {
		return ErrEmptySortBy
	}

	if o.SortOrder != "asc" && o.SortOrder != "desc" {
		return ErrInvalidSortOrder
	}

	if o.Limit < 1 || o.Limit > 100 {
		return ErrInvalidPageSize
	}

	if o.Offset < 1 {
		return ErrInvalidPage
	}

	return nil
}

func (o *FileMetadataListOptions) ValidateEssential() error {
	if o == nil {
		return ErrNilOptions
	}

	if o.UserID == "" {
		return ErrEmptyUserID
	}

	return nil
}

// Validate checks the integrity of the FileMetadataRecord
func (f *FileMetadataRecord) Validate() error {
	if err := f.ValidateEssential(); err != nil {
		return err
	}

	if f.Metadata == nil {
		return ErrNilMetadata
	}

	return nil
}

func (f *FileMetadataRecord) ValidateEssential() error {
	if f == nil {
		return ErrNilRecord
	}

	if f.ID == "" {
		return ErrEmptyID
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

	// ErrNilOptions indicates that the options are nil
	ErrNilOptions = errors.New("options cannot be nil")

	// ErrEmptyUserID indicates that the user ID is empty
	ErrEmptyUserID = errors.New("user ID cannot be empty")

	// ErrEmptySortBy indicates that the sort by is empty
	ErrEmptySortBy = errors.New("sort by cannot be empty")

	// ErrInvalidSortOrder indicates that the sort order is invalid
	ErrInvalidSortOrder = errors.New("sort order must be 'asc' or 'desc'")

	// ErrInvalidPage indicates that the page is invalid
	ErrInvalidPage = errors.New("page must be greater than 0")

	// ErrInvalidPageSize indicates that the page size is invalid
	ErrInvalidPageSize = errors.New("page size must be greater than 0")
)
