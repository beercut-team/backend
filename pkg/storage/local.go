package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type localStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) Storage {
	os.MkdirAll(basePath, 0755)
	return &localStorage{basePath: basePath}
}

func (s *localStorage) Upload(_ context.Context, path string, reader io.Reader, _ int64, _ string) error {
	fullPath := filepath.Join(s.basePath, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	return err
}

func (s *localStorage) Download(_ context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)
	return os.Open(fullPath)
}

func (s *localStorage) Delete(_ context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	return os.Remove(fullPath)
}

func (s *localStorage) PresignedURL(_ context.Context, path string) (string, error) {
	return fmt.Sprintf("/api/v1/media/file/%s", path), nil
}
