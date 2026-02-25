package repository

import (
	"context"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	FindByEntity(ctx context.Context, entity string, entityID uint) ([]domain.AuditLog, error)
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditRepository) FindByEntity(ctx context.Context, entity string, entityID uint) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.WithContext(ctx).Where("entity = ? AND entity_id = ?", entity, entityID).
		Order("created_at DESC").Find(&logs).Error
	return logs, err
}
