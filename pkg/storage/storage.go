package storage

import (
	"context"
	"io"
)

type Storage interface {
	Upload(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	PresignedURL(ctx context.Context, path string) (string, error)
}
