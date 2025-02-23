package local

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	circuit "github.com/yaanno/upload-store-process/services/file-storage-service/internal/breaker"
	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	metadataService "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata"
)

type LocalFileSystem struct {
	basePath        string
	breaker         *circuit.CircuitBreaker
	metadataService metadataService.MetadataService
}

func NewLocalFileSystem(basePath string, metadataService metadataService.MetadataService) *LocalFileSystem {
	return &LocalFileSystem{
		basePath:        basePath,
		breaker:         circuit.NewCircuitBreaker(3, 10*time.Second),
		metadataService: metadataService,
	}
}

func (fs *LocalFileSystem) storeFile(storagePath string, content io.Reader) error {
	var err error

	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(storagePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func (fs *LocalFileSystem) Store(ctx context.Context, fileID string, content io.Reader) (string, error) {
	if err := fs.validateFileID(fileID); err != nil {
		return "", err
	}

	// Create a TeeReader to calculate checksum while writing
	var checksumBuf bytes.Buffer
	teeReader := io.TeeReader(content, &checksumBuf)

	storagePath := filepath.Join(fs.basePath, fileID)
	if fs.fileExists(storagePath) {
		return "", fmt.Errorf("file already exists: %s", fileID)
	}

	err := fs.breaker.Execute(ctx, func() error {
		return fs.storeFile(storagePath, teeReader)
	})
	if err != nil {
		return "", err
	}
	// Calculate and store checksum
	checksum, err := calculateChecksum(&checksumBuf)
	if err != nil {
		return "", err
	}

	metadata := &domain.FileMetadataRecord{
		ID:          fileID,
		Checksum:    checksum,
		StoragePath: storagePath,
	}

	if err := fs.metadataService.UpdateFileMetadata(ctx, fileID, metadata); err != nil {
		return "", fmt.Errorf("failed to update metadata: %w", err)
	}

	return storagePath, nil
}

func (fs *LocalFileSystem) retrieveFile(storagePath string) (*os.File, error) {
	file, err := os.Open(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (fs *LocalFileSystem) Retrieve(ctx context.Context, fileID string) (io.ReadCloser, error) {
	if err := fs.validateFileID(fileID); err != nil {
		return nil, err
	}

	storagePath := filepath.Join(fs.basePath, fileID)

	if !fs.fileExists(storagePath) {
		return nil, fmt.Errorf("file not found: %s", fileID)
	}

	// Verify file integrity before returning
	if err := fs.verifyFileIntegrity(ctx, fileID, storagePath); err != nil {
		return nil, fmt.Errorf("file integrity check failed: %w", err)
	}

	var file *os.File
	err := fs.breaker.Execute(ctx, func() error {
		var err error
		file, err = fs.retrieveFile(storagePath)
		return err
	})
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (fs *LocalFileSystem) deleteFile(storagePath string) error {
	if err := os.Remove(storagePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %w", err)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (fs *LocalFileSystem) Delete(ctx context.Context, fileID string) error {
	if err := fs.validateFileID(fileID); err != nil {
		return err
	}

	storagePath := filepath.Join(fs.basePath, fileID)
	if !fs.fileExists(storagePath) {
		return fmt.Errorf("file not found: %s", fileID)
	}

	return fs.breaker.Execute(ctx, func() error {
		return fs.deleteFile(storagePath)
	})
}

func (fs *LocalFileSystem) listFiles(basePath string) ([]string, error) {
	var files []string
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access path %s: %w", path, err)
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	return files, nil
}

func (fs *LocalFileSystem) List(ctx context.Context) ([]string, error) {
	var files []string
	err := fs.breaker.Execute(ctx, func() error {
		var err error
		files, err = fs.listFiles(fs.basePath)
		return err
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// Add validation helper methods
func (fs *LocalFileSystem) validateFileID(fileID string) error {
	if fileID == "" {
		return fmt.Errorf("file ID cannot be empty")
	}
	if filepath.Clean(fileID) != fileID {
		return fmt.Errorf("invalid file ID path")
	}
	return nil
}

// Add file existence helper
func (fs *LocalFileSystem) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Add checksum calculation helper
func calculateChecksum(file io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// Add file integrity verification helper
func (fs *LocalFileSystem) verifyFileIntegrity(ctx context.Context, fileID, storagePath string) error {
	metadata, err := fs.metadataService.RetrieveFileMetadataByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to retrieve checksum: %w", err)
	}

	file, err := os.Open(storagePath)
	if err != nil {
		return fmt.Errorf("failed to open file for integrity check: %w", err)
	}
	defer file.Close()

	currentChecksum, err := calculateChecksum(file)
	if err != nil {
		return err
	}

	if currentChecksum != metadata.Checksum {
		return fmt.Errorf("file integrity check failed: checksums do not match")
	}

	return nil
}
