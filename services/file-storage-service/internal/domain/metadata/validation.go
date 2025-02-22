package metadata

import "errors"

// Error variables for validation
var (
	ErrNilRecord        = errors.New("file metadata record is nil")
	ErrEmptyID          = errors.New("file metadata record ID cannot be empty")
	ErrNilMetadata      = errors.New("file metadata cannot be nil")
	ErrNilOptions       = errors.New("options cannot be nil")
	ErrEmptyUserID      = errors.New("user ID cannot be empty")
	ErrEmptySortBy      = errors.New("sort by cannot be empty")
	ErrInvalidSortOrder = errors.New("sort order must be 'asc' or 'desc'")
	ErrInvalidPage      = errors.New("page must be greater than 0")
	ErrInvalidPageSize  = errors.New("page size must be greater than 0")
)

// Validate checks the integrity of the FileMetadataListOptions
func (o *FileMetadataListOptions) Validate() error {
	if err := o.ValidateEssential(); err != nil {
		return err
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
