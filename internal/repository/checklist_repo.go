package repository

import (
	"context"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type ChecklistRepository interface {
	CreateTemplate(ctx context.Context, t *domain.ChecklistTemplate) error
	FindTemplatesByOperation(ctx context.Context, opType domain.OperationType) ([]domain.ChecklistTemplate, error)
	CreateItem(ctx context.Context, item *domain.ChecklistItem) error
	CreateItems(ctx context.Context, items []domain.ChecklistItem) error
	FindItemByID(ctx context.Context, id uint) (*domain.ChecklistItem, error)
	FindItemsByPatient(ctx context.Context, patientID uint) ([]domain.ChecklistItem, error)
	UpdateItem(ctx context.Context, item *domain.ChecklistItem) error
	CountByPatient(ctx context.Context, patientID uint) (total int64, completed int64, required int64, requiredCompleted int64, err error)
	FindExpiredItems(ctx context.Context) ([]domain.ChecklistItem, error)
	UpdateItemStatus(ctx context.Context, id uint, status domain.ChecklistItemStatus) error
}

type checklistRepository struct {
	db *gorm.DB
}

func NewChecklistRepository(db *gorm.DB) ChecklistRepository {
	return &checklistRepository{db: db}
}

func (r *checklistRepository) CreateTemplate(ctx context.Context, t *domain.ChecklistTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *checklistRepository) FindTemplatesByOperation(ctx context.Context, opType domain.OperationType) ([]domain.ChecklistTemplate, error) {
	var templates []domain.ChecklistTemplate
	err := r.db.WithContext(ctx).Where("operation_type = ?", opType).Order("sort_order ASC").Find(&templates).Error
	return templates, err
}

func (r *checklistRepository) CreateItem(ctx context.Context, item *domain.ChecklistItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *checklistRepository) CreateItems(ctx context.Context, items []domain.ChecklistItem) error {
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *checklistRepository) FindItemByID(ctx context.Context, id uint) (*domain.ChecklistItem, error) {
	var item domain.ChecklistItem
	if err := r.db.WithContext(ctx).Preload("Template").First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *checklistRepository) FindItemsByPatient(ctx context.Context, patientID uint) ([]domain.ChecklistItem, error) {
	var items []domain.ChecklistItem
	err := r.db.WithContext(ctx).Where("patient_id = ?", patientID).
		Preload("Template").Order("category ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *checklistRepository) UpdateItem(ctx context.Context, item *domain.ChecklistItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *checklistRepository) CountByPatient(ctx context.Context, patientID uint) (total int64, completed int64, required int64, requiredCompleted int64, err error) {
	err = r.db.WithContext(ctx).Model(&domain.ChecklistItem{}).Where("patient_id = ?", patientID).Count(&total).Error
	if err != nil {
		return
	}
	err = r.db.WithContext(ctx).Model(&domain.ChecklistItem{}).Where("patient_id = ? AND status = ?", patientID, domain.ChecklistStatusCompleted).Count(&completed).Error
	if err != nil {
		return
	}
	err = r.db.WithContext(ctx).Model(&domain.ChecklistItem{}).Where("patient_id = ? AND is_required = true", patientID).Count(&required).Error
	if err != nil {
		return
	}
	err = r.db.WithContext(ctx).Model(&domain.ChecklistItem{}).Where("patient_id = ? AND is_required = true AND status = ?", patientID, domain.ChecklistStatusCompleted).Count(&requiredCompleted).Error
	return
}

func (r *checklistRepository) FindExpiredItems(ctx context.Context) ([]domain.ChecklistItem, error) {
	var items []domain.ChecklistItem
	err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < NOW() AND status NOT IN ?",
			[]domain.ChecklistItemStatus{domain.ChecklistStatusExpired, domain.ChecklistStatusCompleted}).
		Find(&items).Error
	return items, err
}

func (r *checklistRepository) UpdateItemStatus(ctx context.Context, id uint, status domain.ChecklistItemStatus) error {
	return r.db.WithContext(ctx).Model(&domain.ChecklistItem{}).Where("id = ?", id).Update("status", status).Error
}
