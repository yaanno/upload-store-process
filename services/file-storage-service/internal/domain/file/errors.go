package file

import "errors"

var (
	// ErrEmptyFileName indicates that the file name is empty
	ErrEmptyFileName = errors.New("file name cannot be empty")

	// ErrInvalidFileSize indicates that the file size is invalid
	ErrInvalidFileSize = errors.New("file size must be greater than 0")

	// ErrEmptyUserID indicates that the user ID is empty
	ErrEmptyUserID = errors.New("user ID cannot be empty")

	// ErrNilContent indicates that the file content is nil
	ErrNilContent = errors.New("file content cannot be nil")

	// ErrFileNotFound indicates that the file was not found
	ErrFileNotFound = errors.New("file not found")

	// ErrInvalidFileState indicates that the file is in an invalid state
	ErrInvalidFileState = errors.New("file is in an invalid state")

	// ErrFileAlreadyExists indicates that the file already exists
	ErrFileAlreadyExists = errors.New("file already exists")

	// ErrFileSizeTooLarge indicates that the file size exceeds the limit
	ErrFileSizeTooLarge = errors.New("file size exceeds the maximum allowed size")

	// ErrInvalidContentType indicates that the content type is not supported
	ErrInvalidContentType = errors.New("content type is not supported")
)
