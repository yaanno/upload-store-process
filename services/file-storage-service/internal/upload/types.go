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
