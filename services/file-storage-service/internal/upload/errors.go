package upload

import "google.golang.org/grpc/codes"

type UploadError struct {
	Code    codes.Code
	Message string
	Err     error
}

func (e *UploadError) Error() string {
	return e.Message
}
