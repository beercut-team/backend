package repository

import (
	"context"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type MediaRepository interface {
	Create(ctx context.Context, media *domain.Media) error
	FindByID(ctx context.Context, id uint) (*domain.Media, error)
	FindByPatient(ctx context.Context, patientID uint) ([]domain.Media, error)
	Delete(ctx context.Context, id uint) error
	FindOrphaned(ctx context.Context) ([]domain.Media, error)
}

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(ctx context.Context, media *domain.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *mediaRepository) FindByID(ctx context.Context, id uint) (*domain.Media, error) {
	var media domain.Media
	if err := r.db.WithContext(ctx).First(&media, id).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) FindByPatient(ctx context.Context, patientID uint) ([]domain.Media, error) {
	var media []domain.Media
	err := r.db.WithContext(ctx).Where("patient_id = ?", patientID).Order("created_at DESC").Find(&media).Error
	return media, err
}

func (r *mediaRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Media{}, id).Error
}

func (r *mediaRepository) FindOrphaned(ctx context.Context) ([]domain.Media, error) {
	var media []domain.Media
	err := r.db.WithContext(ctx).
		Where("patient_id NOT IN (SELECT id FROM patients)").
		Find(&media).Error
	return media, err
}
