package v1

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (req *ListFilesRequest) Validate() error { // Method on the generated struct!
	if req == nil {
		return status.Errorf(codes.InvalidArgument, "request cannot be nil")
	}
	if req.UserId == "" {
		return status.Errorf(codes.InvalidArgument, "user ID is required")
	}
	if req.Page < 1 {
		return status.Errorf(codes.InvalidArgument, "page must be at least 1")
	}
	if req.PageSize <= 0 {
		return status.Errorf(codes.InvalidArgument, "page size must be at least 1")
	}

	return nil
}
