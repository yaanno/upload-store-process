package handlers

import (
	"encoding/json"
	"net/http"

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

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	req := &upload.UploadRequest{
		FileID:             r.FormValue("fileId"),
		StorageUploadToken: r.FormValue("token"),
		FileContent:        r.Body,
		UserID:             r.FormValue("userId"),
	}

	resp, err := h.uploadService.Upload(r.Context(), req)
	if err != nil {
		// h.handleError(w, err)
		return
	}

	json.NewEncoder(w).Encode(resp)
}
