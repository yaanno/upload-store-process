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

func (fs *LocalFileSystem) List(ctx context.Context) ([]string, error) {
	var files []string
	err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(fs.basePath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (fs *LocalFileSystem) GetBasePath() string {
	return fs.basePath
}

func (fs *LocalFileSystem) GenerateStoragePath(fileID, fileName string) string {
	return ""
}
