package handlers

import (
	"bytes"
	"context"

	uploadv1 "github.com/yaanno/upload-store-process/gen/go/fileupload/v1"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/upload"
)

type UploadHandler struct {
	uploadService upload.UploadService
}

func NewUploadHandler(uploadService upload.UploadService) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
	}
}

func (h *UploadHandler) UploadFile(ctx context.Context, req *uploadv1.UploadFileRequest) (*uploadv1.UploadFileResponse, error) {
	resp, err := h.uploadService.Upload(ctx, &upload.UploadRequest{
		FileID:             req.FileId,
		StorageUploadToken: req.StorageUploadToken,
		FileContent:        bytes.NewReader(req.FileContent),
		UserID:             req.UserId,
	})
	if err != nil {
		return nil, err
	}

	return &uploadv1.UploadFileResponse{
		FileId:      resp.FileID,
		StoragePath: resp.StoragePath,
		BaseResponse: &sharedv1.Response{
			Message: resp.Message,
		},
	}, nil
}
