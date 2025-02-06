package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalFilesystemStorage implements FileStorageProvider for local filesystem
type LocalFilesystemStorage struct {
	BaseUploadPath string
}

// NewLocalFilesystemStorage creates a new local filesystem storage provider
func NewLocalFilesystemStorage(baseUploadPath string) *LocalFilesystemStorage {
	return &LocalFilesystemStorage{
		BaseUploadPath: baseUploadPath,
	}
}

// StoreFile stores a file in the local filesystem
func (lfs *LocalFilesystemStorage) StoreFile(
	ctx context.Context, 
	fileID string, 
	originalFilename string, 
	fileReader io.Reader,
) (string, error) {
	// Generate storage path
	storagePath := lfs.GenerateStoragePath(fileID, originalFilename)
	
	// Create full path
	fullPath := filepath.Join(lfs.BaseUploadPath, storagePath)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Create destination file
	destFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer destFile.Close()
	
	// Copy file contents
	if _, err := io.Copy(destFile, fileReader); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	
	return storagePath, nil
}

// RetrieveFile retrieves a file from local filesystem
func (lfs *LocalFilesystemStorage) RetrieveFile(
	ctx context.Context, 
	storagePath string,
) (io.Reader, error) {
	fullPath := filepath.Join(lfs.BaseUploadPath, storagePath)
	
	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	
	return file, nil
}

// DeleteFile removes a file from local filesystem
func (lfs *LocalFilesystemStorage) DeleteFile(
	ctx context.Context, 
	storagePath string,
) error {
	fullPath := filepath.Join(lfs.BaseUploadPath, storagePath)
	
	// Remove file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// GenerateStoragePath creates a consistent path for file storage
func (lfs *LocalFilesystemStorage) GenerateStoragePath(
	fileID string, 
	originalFilename string,
) string {
	now := time.Now()
	return filepath.Join(
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		fmt.Sprintf("%s_%s", fileID, originalFilename),
	)
}
