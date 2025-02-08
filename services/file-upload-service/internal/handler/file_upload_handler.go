package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	uploadv1 "github.com/yaanno/upload-store-process/gen/go/fileupload/v1"

	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

type FileUploadHandler struct {
	logger      logger.Logger
	maxFileSize int64
	service     uploadv1.FileUploadServiceClient
}

func NewFileUploadHandler(logger logger.Logger, maxFileSize int64, service uploadv1.FileUploadServiceClient) *FileUploadHandler {
	return &FileUploadHandler{logger: logger, maxFileSize: maxFileSize, service: service}
}

// TODO: implement the prepareUpload method
// TODO: implement the cancelUpload method
// TODO: implement the listFiles method

func (h *FileUploadHandler) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
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

	// fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
	// if err != nil {
	// 	h.logger.Error().Err(err).Str("field", "file_size").Msg("Invalid file size")
	// 	http.Error(w, "Invalid file size", http.StatusBadRequest)
	// 	return
	// }

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
