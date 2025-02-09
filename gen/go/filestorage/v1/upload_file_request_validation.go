package v1

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (req *UploadFileRequest) Validate() error { // Method on the generated struct!
	if req == nil {
		return status.Errorf(codes.InvalidArgument, "request cannot be nil")
	}
	if req.UserId == "" {
		return status.Errorf(codes.InvalidArgument, "user ID is required")
	}
	if req.StorageUploadToken == "" {
		return status.Errorf(codes.InvalidArgument, "storage upload token is required")
	}

	return nil
}
