package local

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type LocalFileSystem struct {
	basePath string
}

func NewLocalFileSystem(basePath string) *LocalFileSystem {
	return &LocalFileSystem{
		basePath: basePath,
	}
}

func (fs *LocalFileSystem) Store(ctx context.Context, fileID string, content io.Reader) (string, error) {
	storagePath := filepath.Join(fs.basePath, fileID)

	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return "", err
	}

	f, err := os.Create(storagePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, content); err != nil {
		return "", err
	}

	return storagePath, nil
}

func (fs *LocalFileSystem) Retrieve(ctx context.Context, fileID string) (io.ReadCloser, error) {
	storagePath := filepath.Join(fs.basePath, fileID)
	return os.Open(storagePath)
}

func (fs *LocalFileSystem) Delete(ctx context.Context, fileID string) error {
	storagePath := filepath.Join(fs.basePath, fileID)
	return os.Remove(storagePath)
}
