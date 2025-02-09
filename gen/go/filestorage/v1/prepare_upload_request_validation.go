package v1

import (
	"path/filepath"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (req *PrepareUploadRequest) Validate() error { // Method on the generated struct!
	if req == nil {
		return status.Errorf(codes.InvalidArgument, "request cannot be nil")
	}
	if req.UserId == "" {
		return status.Errorf(codes.InvalidArgument, "user ID is required")
	}
	if req.Filename == "" {
		return status.Errorf(codes.InvalidArgument, "filename cannot be empty")
	}
	if req.FileSizeBytes <= 0 {
		return status.Errorf(codes.InvalidArgument, "filesize cannot be empty")
	}
	if req.FileSizeBytes > 500*1024*1024 { // 500 MB
		return status.Errorf(codes.InvalidArgument, "file too large")
	}
	if !isAllowedFileType(req.Filename) {
		return status.Errorf(codes.InvalidArgument, "unsupported file type")
	}
	return nil
}

func isAllowedFileType(filename string) bool {
	allowedExtensions := []string{".csv", ".json", ".txt"}
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}
