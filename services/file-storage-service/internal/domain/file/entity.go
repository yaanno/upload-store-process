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
