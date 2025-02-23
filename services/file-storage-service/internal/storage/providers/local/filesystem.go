package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	circuit "github.com/yaanno/upload-store-process/services/file-storage-service/internal/breaker"
)

type LocalFileSystem struct {
	basePath string
	breaker  *circuit.CircuitBreaker
}

func NewLocalFileSystem(basePath string) *LocalFileSystem {
	return &LocalFileSystem{
		basePath: basePath,
		breaker:  circuit.NewCircuitBreaker(3, 10*time.Second),
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

	storagePath := filepath.Join(fs.basePath, fileID)
	if fs.fileExists(storagePath) {
		return "", fmt.Errorf("file already exists: %s", fileID)
	}

	err := fs.breaker.Execute(ctx, func() error {
		return fs.storeFile(storagePath, content)
	})
	if err != nil {
		return "", err
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

func (fs *LocalFileSystem) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
