package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	uploadv1 "github.com/yaanno/upload-store-process/gen/go/fileupload/v1"

	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

type FileUploadHandler interface {
	HandleFileUpload(w http.ResponseWriter, r *http.Request)
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

// TODO: implement the prepareUpload method
// TODO: implement the cancelUpload method
// TODO: implement the listFiles method

func (h *FileUploadHandlerImpl) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Validate request method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(h.maxFileSize); err != nil {
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
	file, handler, err := r.FormFile("file")
	if err != nil {
		h.logger.Error().Err(err).Msg("No file uploaded")
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileContentBuffer := bytes.NewBuffer(nil)
	buffer := make([]byte, 4*1024)
	_, err = io.CopyBuffer(fileContentBuffer, io.LimitReader(file, h.maxFileSize), buffer)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read file")
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	fileContent := fileContentBuffer.Bytes()

	grpcRequest := &uploadv1.UploadFileRequest{
		FileId:             fileId,
		FileContent:        fileContent,
		StorageUploadToken: storageUploadToken,
	}

	// Call gRPC service
	response, err := h.service.UploadFile(ctx, grpcRequest)
	if err != nil {
		h.logger.Error().Err(err).Msg("gRPC upload failed")
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}

	// Log successful upload
	h.logger.Info().
		Str("path", response.GetStoragePath()).
		Str("fileName", handler.Filename).
		Msg("File uploaded successfully")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"path":    response.GetStoragePath(),
		"message": response.GetBaseResponse().Message,
	})
}

// TODO: add fileID and StorageUploadToken for validation
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
