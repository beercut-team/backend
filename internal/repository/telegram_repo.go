package repository

import (
	"context"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type TelegramRepository interface {
	Create(ctx context.Context, binding *domain.TelegramBinding) error
	FindByChatID(ctx context.Context, chatID int64) (*domain.TelegramBinding, error)
	FindByPatientID(ctx context.Context, patientID uint) (*domain.TelegramBinding, error)
	Delete(ctx context.Context, chatID int64) error
}

type telegramRepository struct {
	db *gorm.DB
}

func NewTelegramRepository(db *gorm.DB) TelegramRepository {
	return &telegramRepository{db: db}
}

func (r *telegramRepository) Create(ctx context.Context, binding *domain.TelegramBinding) error {
	return r.db.WithContext(ctx).Create(binding).Error
}

func (r *telegramRepository) FindByChatID(ctx context.Context, chatID int64) (*domain.TelegramBinding, error) {
	var binding domain.TelegramBinding
	if err := r.db.WithContext(ctx).Where("chat_id = ? AND is_active = true", chatID).First(&binding).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

func (r *telegramRepository) FindByPatientID(ctx context.Context, patientID uint) (*domain.TelegramBinding, error) {
	var binding domain.TelegramBinding
	if err := r.db.WithContext(ctx).Where("patient_id = ? AND is_active = true", patientID).First(&binding).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

func (r *telegramRepository) Delete(ctx context.Context, chatID int64) error {
	return r.db.WithContext(ctx).Model(&domain.TelegramBinding{}).Where("chat_id = ?", chatID).Update("is_active", false).Error
}
