package upload

import "io"

type UploadRequest struct {
	FileID             string
	StorageUploadToken string
	FileSizeBytes      int64
	FileContent        io.Reader
	UserID             string
}

type UploadResponse struct {
	FileID      string
	StoragePath string
	Message     string
}

type PrepareUploadRequest struct {
	Filename      string
	FileSizeBytes int64
	UserID        string
	ContentType   string
}

type PrepareUploadResponse struct {
	FileID      string
	UploadToken string
	StoragePath string
	ExpiresAt   int64
	Message     string
}
