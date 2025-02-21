package storage

import "errors"

var (
	ErrFileNotFound  = errors.New("file not found in storage")
	ErrStorageFull   = errors.New("storage capacity exceeded")
	ErrInvalidPath   = errors.New("invalid storage path")
	ErrStorageAccess = errors.New("storage access denied")
)
