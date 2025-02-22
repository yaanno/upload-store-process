package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	gatewayv1 "github.com/yaanno/upload-store-process/gen/go/gateway/v1"

	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

type FileUploadHandler interface {
	// HandleFileUpload(w http.ResponseWriter, r *http.Request)
	PrepareUpload(w http.ResponseWriter, r *http.Request)
	ListFiles(w http.ResponseWriter, r *http.Request)
	DeleteFile(w http.ResponseWriter, r *http.Request)
	GetFileStatus(w http.ResponseWriter, r *http.Request)
	GetFileMetadata(w http.ResponseWriter, r *http.Request)
}

type FileUploadHandlerImpl struct {
	logger      logger.Logger
	maxFileSize int64
	service     gatewayv1.APIGatewayServiceClient
}

func NewFileUploadHandler(logger logger.Logger, maxFileSize int64, service gatewayv1.APIGatewayServiceClient) *FileUploadHandlerImpl {
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

	grpcRequest := &gatewayv1.PrepareUploadRequest{
		Filename:      req.Filename,
		FileSizeBytes: req.FileSizeBytes,
		ContentType:   req.FileType,
		UserId:        "1", // TODO: get user ID from JWT
	}

	response, err := h.service.PrepareUpload(ctx, grpcRequest)
	if err != nil {
		h.logger.Error().Err(err).Msg("gRPC prepareupload failed")
	}

	// Log successful upload
	h.logger.Info().
		Str("file_id", response.GetFileId()).
		Str("upload_token", response.GetUploadToken()).
		Msg("File uploaded successfully")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"storage_upload_token": response.GetUploadToken(),
		"message":              response.GetBaseResponse().Message,
		"expiration_time":      response.GetExpiresAt().AsTime().Format(time.RFC3339),
		"file_id":              response.GetFileId(),
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

	grpcRequest := &gatewayv1.GetFileMetadataRequest{
		FileId: fileId,
		UserId: "1", // TODO: get user ID from JWT
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

func (h *FileUploadHandlerImpl) ListFiles(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	grpcRequest := &gatewayv1.ListFilesRequest{
		UserId: "1", // TODO: get user ID from JWT
	}

	response, err := h.service.ListFiles(ctx, grpcRequest)
	if err != nil {
		h.logger.Error().Err(err).Msg("gRPC list files failed")
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	// Log successful upload
	h.logger.Info().
		Msg("files listed successfully")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.Files)
}

func (h *FileUploadHandlerImpl) DeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var fileId string
	err := json.NewDecoder(r.Body).Decode(&fileId)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read file ID")
		http.Error(w, "Failed to read file ID", http.StatusBadRequest)
		return
	}
	grpcRequest := &gatewayv1.DeleteFileRequest{
		FileId:      fileId,
		ForceDelete: false,
		UserId:      "1", // TODO: get user ID from JWT
	}
	response, err := h.service.DeleteFile(ctx, grpcRequest)
	if err != nil {
		h.logger.Error().Err(err).Msg("gRPC delete file failed")
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	// Log successful upload
	h.logger.Info().
		Msg("file deleted successfully")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *FileUploadHandlerImpl) GetFileStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var fileId string
	err := json.NewDecoder(r.Body).Decode(&fileId)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read file ID")
		http.Error(w, "Failed to read file ID", http.StatusBadRequest)
		return
	}
	grpcRequest := &gatewayv1.GetFileStatusRequest{
		FileId: fileId,
		UserId: "1", // TODO: get user ID from JWT,
	}
	response, err := h.service.GetFileStatus(ctx, grpcRequest)
	if err != nil {
		h.logger.Error().Err(err).Msg("gRPC get file status failed")
		http.Error(w, "Failed to get file status", http.StatusInternalServerError)
		return
	}

	// Log successful upload
	h.logger.Info().
		Msg("file status retrieved successfully")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var _ FileUploadHandler = (*FileUploadHandlerImpl)(nil)

//
