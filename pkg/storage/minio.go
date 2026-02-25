package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

type minioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinIOStorage(cfg *config.Config) (Storage, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось создать MinIO клиент: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinIOBucket)
	if err != nil {
		return nil, fmt.Errorf("не удалось проверить bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("не удалось создать bucket: %w", err)
		}
		log.Info().Str("bucket", cfg.MinIOBucket).Msg("создан MinIO bucket")
	}

	return &minioStorage{client: client, bucket: cfg.MinIOBucket}, nil
}

func (s *minioStorage) Upload(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, path, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *minioStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
}

func (s *minioStorage) Delete(ctx context.Context, path string) error {
	return s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
}

func (s *minioStorage) PresignedURL(ctx context.Context, path string) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, path, 15*time.Minute, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
