package repository

import (
	"context"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type SurgeryRepository interface {
	Create(ctx context.Context, surgery *domain.Surgery) error
	FindByID(ctx context.Context, id uint) (*domain.Surgery, error)
	FindByPatient(ctx context.Context, patientID uint) ([]domain.Surgery, error)
	FindBySurgeon(ctx context.Context, surgeonID uint, offset, limit int) ([]domain.Surgery, int64, error)
	Update(ctx context.Context, surgery *domain.Surgery) error
	FindUpcoming(ctx context.Context, before time.Time) ([]domain.Surgery, error)
}

type surgeryRepository struct {
	db *gorm.DB
}

func NewSurgeryRepository(db *gorm.DB) SurgeryRepository {
	return &surgeryRepository{db: db}
}

func (r *surgeryRepository) Create(ctx context.Context, surgery *domain.Surgery) error {
	return r.db.WithContext(ctx).Create(surgery).Error
}

func (r *surgeryRepository) FindByID(ctx context.Context, id uint) (*domain.Surgery, error) {
	var surgery domain.Surgery
	if err := r.db.WithContext(ctx).Preload("Patient").Preload("Surgeon").First(&surgery, id).Error; err != nil {
		return nil, err
	}
	return &surgery, nil
}

func (r *surgeryRepository) FindByPatient(ctx context.Context, patientID uint) ([]domain.Surgery, error) {
	var surgeries []domain.Surgery
	err := r.db.WithContext(ctx).Where("patient_id = ?", patientID).Preload("Surgeon").Order("scheduled_date DESC").Find(&surgeries).Error
	return surgeries, err
}

func (r *surgeryRepository) FindBySurgeon(ctx context.Context, surgeonID uint, offset, limit int) ([]domain.Surgery, int64, error) {
	var surgeries []domain.Surgery
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Surgery{}).Where("surgeon_id = ?", surgeonID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Preload("Patient").Offset(offset).Limit(limit).Order("scheduled_date ASC").Find(&surgeries).Error; err != nil {
		return nil, 0, err
	}

	return surgeries, total, nil
}

func (r *surgeryRepository) Update(ctx context.Context, surgery *domain.Surgery) error {
	return r.db.WithContext(ctx).Save(surgery).Error
}

func (r *surgeryRepository) FindUpcoming(ctx context.Context, before time.Time) ([]domain.Surgery, error) {
	var surgeries []domain.Surgery
	err := r.db.WithContext(ctx).
		Where("scheduled_date <= ? AND scheduled_date > NOW() AND status = ?", before, domain.SurgeryStatusScheduled).
		Preload("Patient").Preload("Surgeon").
		Find(&surgeries).Error
	return surgeries, err
}
