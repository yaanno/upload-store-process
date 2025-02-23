package file

import (
	"io"
	"time"
)

// File represents a file in the system
type File struct {
	ID              string
	Name            string
	Size            int64
	ContentType     string
	Content         io.Reader
	UploadedAt      time.Time
	UserID          string
	ProcessingState string
}

// FileStatus represents the current state of a file
type FileStatus string

const (
	StatusPending   FileStatus = "PENDING"
	StatusUploading FileStatus = "UPLOADING"
	StatusComplete  FileStatus = "COMPLETE"
	StatusFailed    FileStatus = "FAILED"
)

// NewFile creates a new File instance
func NewFile(name string, size int64, contentType string, content io.Reader, userID string) *File {
	return &File{
		Name:            name,
		Size:            size,
		ContentType:     contentType,
		Content:         content,
		UserID:          userID,
		ProcessingState: string(StatusPending),
		UploadedAt:      time.Now().UTC(),
	}
}

// Validate checks if the file meets basic requirements
func (f *File) Validate() error {
	if f.Name == "" {
		return ErrEmptyFileName
	}
	if f.Size <= 0 {
		return ErrInvalidFileSize
	}
	if f.UserID == "" {
		return ErrEmptyUserID
	}
	if f.Content == nil {
		return ErrNilContent
	}
	return nil
}

// SetStatus updates the processing state of the file
func (f *File) SetStatus(status FileStatus) {
	f.ProcessingState = string(status)
}

// IsComplete checks if the file processing is complete
func (f *File) IsComplete() bool {
	return f.ProcessingState == string(StatusComplete)
}

// IsFailed checks if the file processing has failed
func (f *File) IsFailed() bool {
	return f.ProcessingState == string(StatusFailed)
}

// IsPending checks if the file processing is pending
func (f *File) IsPending() bool {
	return f.ProcessingState == string(StatusPending)
}

// IsUploading checks if the file is currently being uploaded
func (f *File) IsUploading() bool {
	return f.ProcessingState == string(StatusUploading)
}

// GetStatus returns the current processing state of the file
func (f *File) GetStatus() FileStatus {
	return FileStatus(f.ProcessingState)
}
