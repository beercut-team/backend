package repository

import (
	"context"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type IOLRepository interface {
	Create(ctx context.Context, calc *domain.IOLCalculation) error
	FindByPatient(ctx context.Context, patientID uint) ([]domain.IOLCalculation, error)
	FindByID(ctx context.Context, id uint) (*domain.IOLCalculation, error)
}

type iolRepository struct {
	db *gorm.DB
}

func NewIOLRepository(db *gorm.DB) IOLRepository {
	return &iolRepository{db: db}
}

func (r *iolRepository) Create(ctx context.Context, calc *domain.IOLCalculation) error {
	return r.db.WithContext(ctx).Create(calc).Error
}

func (r *iolRepository) FindByPatient(ctx context.Context, patientID uint) ([]domain.IOLCalculation, error) {
	var calcs []domain.IOLCalculation
	err := r.db.WithContext(ctx).Where("patient_id = ?", patientID).Order("created_at DESC").Find(&calcs).Error
	return calcs, err
}

func (r *iolRepository) FindByID(ctx context.Context, id uint) (*domain.IOLCalculation, error) {
	var calc domain.IOLCalculation
	if err := r.db.WithContext(ctx).First(&calc, id).Error; err != nil {
		return nil, err
	}
	return &calc, nil
}
