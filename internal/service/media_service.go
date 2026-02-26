package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"github.com/beercut-team/backend-boilerplate/pkg/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxFileSize = 20 * 1024 * 1024 // 20MB

type MediaService interface {
	Upload(ctx context.Context, patientID, uploadedBy uint, fileName, contentType, category string, size int64, reader io.Reader) (*domain.Media, error)
	GetByID(ctx context.Context, id uint) (*domain.Media, error)
	GetByPatient(ctx context.Context, patientID uint) ([]domain.Media, error)
	Delete(ctx context.Context, id uint) error
	DownloadFile(ctx context.Context, id uint) (io.ReadCloser, string, error)
	GetDownloadURL(ctx context.Context, id uint) (string, error)
	GetThumbnailURL(ctx context.Context, id uint) (string, error)
}

type mediaService struct {
	repo    repository.MediaRepository
	storage storage.Storage
}

func NewMediaService(repo repository.MediaRepository, store storage.Storage) MediaService {
	return &mediaService{repo: repo, storage: store}
}

func (s *mediaService) Upload(ctx context.Context, patientID, uploadedBy uint, fileName, contentType, category string, size int64, reader io.Reader) (*domain.Media, error) {
	if size > maxFileSize {
		return nil, errors.New("файл слишком большой, максимум 20МБ")
	}

	ext := filepath.Ext(fileName)
	uid := uuid.New().String()
	storagePath := fmt.Sprintf("%d/%s/%s%s", patientID, category, uid, ext)

	if err := s.storage.Upload(ctx, storagePath, reader, size, contentType); err != nil {
		return nil, fmt.Errorf("не удалось загрузить файл: %w", err)
	}

	// Generate thumbnail for images
	var thumbPath string
	if strings.HasPrefix(contentType, "image/") {
		thumbPath = fmt.Sprintf("%d/%s/%s_thumb%s", patientID, category, uid, ext)
	}

	media := &domain.Media{
		PatientID:     patientID,
		UploadedBy:    uploadedBy,
		FileName:      uid + ext,
		OriginalName:  fileName,
		ContentType:   contentType,
		Size:          size,
		StoragePath:   storagePath,
		ThumbnailPath: thumbPath,
		Category:      category,
	}

	if err := s.repo.Create(ctx, media); err != nil {
		return nil, errors.New("не удалось сохранить запись медиа")
	}

	return media, nil
}

func (s *mediaService) GetByID(ctx context.Context, id uint) (*domain.Media, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("медиафайл не найден")
		}
		return nil, err
	}
	return m, nil
}

func (s *mediaService) GetByPatient(ctx context.Context, patientID uint) ([]domain.Media, error) {
	return s.repo.FindByPatient(ctx, patientID)
}

func (s *mediaService) Delete(ctx context.Context, id uint) error {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("медиафайл не найден")
	}

	s.storage.Delete(ctx, m.StoragePath)
	if m.ThumbnailPath != "" {
		s.storage.Delete(ctx, m.ThumbnailPath)
	}

	return s.repo.Delete(ctx, id)
}

func (s *mediaService) DownloadFile(ctx context.Context, id uint) (io.ReadCloser, string, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", errors.New("медиафайл не найден")
	}

	reader, err := s.storage.Download(ctx, m.StoragePath)
	if err != nil {
		return nil, "", fmt.Errorf("не удалось скачать файл: %w", err)
	}

	return reader, m.ContentType, nil
}

func (s *mediaService) GetDownloadURL(ctx context.Context, id uint) (string, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", errors.New("медиафайл не найден")
	}
	return s.storage.PresignedURL(ctx, m.StoragePath)
}

func (s *mediaService) GetThumbnailURL(ctx context.Context, id uint) (string, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", errors.New("медиафайл не найден")
	}
	if m.ThumbnailPath == "" {
		return "", errors.New("миниатюра недоступна")
	}
	return s.storage.PresignedURL(ctx, m.ThumbnailPath)
}
