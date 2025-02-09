package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	uploadv1 "github.com/yaanno/upload-store-process/gen/go/fileupload/v1"

	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

type FileUploadHandler interface {
	// HandleFileUpload(w http.ResponseWriter, r *http.Request)
	PrepareUpload(w http.ResponseWriter, r *http.Request)
	CancelUpload(w http.ResponseWriter, r *http.Request)
	// ListFiles(w http.ResponseWriter, r *http.Request)
	// GetFile(w http.ResponseWriter, r *http.Request)
	// DeleteFile(w http.ResponseWriter, r *http.Request)
	// DownloadFile(w http.ResponseWriter, r *http.Request)
	GetFileMetadata(w http.ResponseWriter, r *http.Request)
	// UpdateFileMetadata(w http.ResponseWriter, r *http.Request)

}

type FileUploadHandlerImpl struct {
	logger      logger.Logger
	maxFileSize int64
	service     uploadv1.FileUploadServiceClient
}

func NewFileUploadHandler(logger logger.Logger, maxFileSize int64, service uploadv1.FileUploadServiceClient) *FileUploadHandlerImpl {
	return &FileUploadHandlerImpl{logger: logger, maxFileSize: maxFileSize, service: service}
}

func (h *FileUploadHandlerImpl) PrepareUpload(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	type Request struct {
		Filename      string `json:"filename"`
		FileSizeBytes int64  `json:"file_size"`
		FileType      string `json:"file_type"`
	}

	var req Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {

	}

	grpcRequest := &uploadv1.PrepareUploadRequest{
		Filename:      req.Filename,
		FileSizeBytes: req.FileSizeBytes,
		FileType:      req.FileType,
	}

	response, err := h.service.PrepareUpload(ctx, grpcRequest)
	if err != nil {
		h.logger.Error().Err(err).Msg("gRPC prepareupload failed")
	}

	// Log successful upload
	h.logger.Info().
		Str("path", response.GetStoragePath()).
		Str("upload_token", response.StorageUploadToken).
		Msg("File uploaded successfully")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"storage_path":         response.GetStoragePath(),
		"storage_upload_token": response.GetStorageUploadToken(),
		"message":              response.GetBaseResponse().Message,
		"expiration_time":      strconv.FormatInt(response.GetExpirationTime(), 10),
		"global_upload_id":     response.GetGlobalUploadId(),
	})
}

// TODO: REMOVE - MOVE TO THE NEW OBJECT STORAGE SERVICE

// TODO: REMOVE COMPLETELY FROM THE CODEBASE
func (h *FileUploadHandlerImpl) CancelUpload(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var uploadID string
	err := json.NewDecoder(r.Body).Decode(&uploadID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read file ID")
		http.Error(w, "Failed to read file ID", http.StatusBadRequest)
		return
	}

	grpcRequest := &uploadv1.CancelUploadRequest{
		GlobalUploadId: uploadID,
		Reason:         "User cancelled the upload",
	}
	response, err := h.service.CancelUpload(ctx, grpcRequest)
	if err != nil {
		h.logger.Error().Err(err).Msg("upload cancellation failed")
	}

	// Log successful upload
	h.logger.Info().
		Msg("User cancelled the upload successfully")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":   response.GetBaseResponse().Message,
		"cancelled": strconv.FormatBool(response.GetUploadCancelled()),
	})

}

func (h *FileUploadHandlerImpl) GetFileMetadata(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var fileId string
	err := json.NewDecoder(r.Body).Decode(&fileId)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read file ID")
		http.Error(w, "Failed to read file ID", http.StatusBadRequest)
		return
	}

	grpcRequest := &uploadv1.GetFileMetadataRequest{
		FileId: fileId,
	}

	response, err := h.service.GetFileMetadata(ctx, grpcRequest)
	if err != nil {
		h.logger.Error().Err(err).Msg("gRPC get file metadata failed")
		http.Error(w, "Failed to get file metadata", http.StatusInternalServerError)
		return
	}

	// Log successful upload
	h.logger.Info().
		Msg("metadata retrieved successfully")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.Metadata)

}

var _ FileUploadHandler = (*FileUploadHandlerImpl)(nil)

//
