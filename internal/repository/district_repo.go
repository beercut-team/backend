package repository

import (
	"context"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type DistrictRepository interface {
	Create(ctx context.Context, district *domain.District) error
	FindByID(ctx context.Context, id uint) (*domain.District, error)
	FindAll(ctx context.Context, search string, offset, limit int) ([]domain.District, int64, error)
	Update(ctx context.Context, district *domain.District) error
	Delete(ctx context.Context, id uint) error
}

type districtRepository struct {
	db *gorm.DB
}

func NewDistrictRepository(db *gorm.DB) DistrictRepository {
	return &districtRepository{db: db}
}

func (r *districtRepository) Create(ctx context.Context, district *domain.District) error {
	return r.db.WithContext(ctx).Create(district).Error
}

func (r *districtRepository) FindByID(ctx context.Context, id uint) (*domain.District, error) {
	var district domain.District
	if err := r.db.WithContext(ctx).First(&district, id).Error; err != nil {
		return nil, err
	}
	return &district, nil
}

func (r *districtRepository) FindAll(ctx context.Context, search string, offset, limit int) ([]domain.District, int64, error) {
	var districts []domain.District
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.District{})
	if search != "" {
		query = query.Where("name ILIKE ? OR region ILIKE ? OR code ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("name ASC").Find(&districts).Error; err != nil {
		return nil, 0, err
	}

	return districts, total, nil
}

func (r *districtRepository) Update(ctx context.Context, district *domain.District) error {
	return r.db.WithContext(ctx).Save(district).Error
}

func (r *districtRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.District{}, id).Error
}
