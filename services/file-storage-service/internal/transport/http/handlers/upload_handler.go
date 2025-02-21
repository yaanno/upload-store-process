package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	service "github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB
)

type FileUploadHandler interface {
	CreateFile(w http.ResponseWriter, r *http.Request)
	GetFile(w http.ResponseWriter, r *http.Request)
	DeleteFile(w http.ResponseWriter, r *http.Request)
}

type FileUploadHandlerImpl struct {
	logger  logger.Logger
	service service.FileUploadService
}

func NewFileUploadHandler(logger logger.Logger, service service.FileUploadService) *FileUploadHandlerImpl {
	return &FileUploadHandlerImpl{logger: logger, service: service}
}

func (h *FileUploadHandlerImpl) CreateFile(w http.ResponseWriter, r *http.Request) {
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

	// Lightweight service call
	// update the metadata in the repository
	// return with metadata
	err = h.service.UploadFile(ctx, &service.UploadFileRequest{
		FileId:             fileId,
		StorageUploadToken: storageUploadToken,
		FileSizeBytes:      fileSizeStr,
		FileContent:        fileContentBuffer,
	})
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to upload file")
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	// TODO: decide if the service should handle this or keep it lightweight
	// UploadService should handle this
	// it just might return the error and the metadata after initially creating one

	// Store file
	// storagePath, err := h.store.StoreFile(ctx, fileId, "", fileContentBuffer)
	// if err != nil {
	// 	h.logger.Error().Err(err).Msg("Failed to store file")
	// 	http.Error(w, "Failed to store file", http.StatusInternalServerError)
	// 	return
	// }

	// TODO: Save metadata

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"storage_path": "",
	})
	w.WriteHeader(http.StatusCreated)
}

func (h *FileUploadHandlerImpl) GetFile(w http.ResponseWriter, r *http.Request) {
}

func (h *FileUploadHandlerImpl) DeleteFile(w http.ResponseWriter, r *http.Request) {
}

var _ FileUploadHandler = (*FileUploadHandlerImpl)(nil)
