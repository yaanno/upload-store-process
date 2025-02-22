package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/upload"
	service "github.com/yaanno/upload-store-process/services/file-storage-service/internal/upload"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB
)

type UploadHandler interface {
	CreateFile(w http.ResponseWriter, r *http.Request)
	GetFile(w http.ResponseWriter, r *http.Request)
	DeleteFile(w http.ResponseWriter, r *http.Request)
	Upload(w http.ResponseWriter, r *http.Request)
}

type UploadHandlerImpl struct {
	logger        logger.Logger
	uploadService service.UploadService
}

func NewFileUploadHandler(logger logger.Logger, uploadService service.UploadService) *UploadHandlerImpl {
	return &UploadHandlerImpl{logger: logger, uploadService: uploadService}
}

func (h *UploadHandlerImpl) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileSizeStr := r.Header.Get("Content-Length")
	fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid file size", http.StatusBadRequest)
		return
	}

	req := &upload.UploadRequest{
		FileID:             r.FormValue("fileId"),
		StorageUploadToken: r.FormValue("token"),
		FileSizeBytes:      fileSize,
		FileContent:        r.Body,
		UserID:             r.FormValue("userId"),
	}

	resp, err := h.uploadService.Upload(r.Context(), req)
	if err != nil {
		// statusCode, errorResponse := errors.MapToHTTPError(err)
		// w.WriteHeader(statusCode)
		// json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UploadHandlerImpl) CreateFile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Validate request method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse multipart form")
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	fileId := r.FormValue("file_id")
	if fileId == "" {
		h.logger.Error().Str("field", "file_id").Msg("File ID cannot be empty")
		http.Error(w, "File ID cannot be empty", http.StatusBadRequest)
		return
	}

	// TODO: validate storage upload token
	storageUploadToken := r.FormValue("storage_upload_token")
	if storageUploadToken == "" {
		h.logger.Error().Str("field", "storage_upload_token").Msg("Storage upload token cannot be empty")
		http.Error(w, "Storage upload token cannot be empty", http.StatusBadRequest)
		return
	}

	fileSizeStr := r.FormValue("file_size")
	if fileSizeStr == "" {
		h.logger.Error().Str("field", "file_size").Msg("File size cannot be empty")
		http.Error(w, "File size cannot be empty", http.StatusBadRequest)
		return
	}

	// Extract file from request
	file, _, err := r.FormFile("file")
	if err != nil {
		h.logger.Error().Err(err).Msg("No file uploaded")
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileContentBuffer := bytes.NewBuffer(nil)
	buffer := make([]byte, 4*1024)
	_, err = io.CopyBuffer(fileContentBuffer, io.LimitReader(file, maxFileSize), buffer)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read file")
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	resp, err := h.uploadService.Upload(ctx, &service.UploadRequest{
		FileID:             fileId,
		StorageUploadToken: storageUploadToken,
		// FileSizeBytes:      fileSizeStr,
		FileContent: fileContentBuffer,
	})
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to upload file")
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&resp)
	w.WriteHeader(http.StatusCreated)
}

func (h *UploadHandlerImpl) GetFile(w http.ResponseWriter, r *http.Request) {
}

func (h *UploadHandlerImpl) DeleteFile(w http.ResponseWriter, r *http.Request) {
}

var _ UploadHandler = (*UploadHandlerImpl)(nil)
