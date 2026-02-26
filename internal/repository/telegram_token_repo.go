package repository

import (
	"context"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type TelegramTokenRepository interface {
	Create(ctx context.Context, token *domain.TelegramLoginToken) error
	FindByToken(ctx context.Context, token string) (*domain.TelegramLoginToken, error)
	MarkAsUsed(ctx context.Context, tokenID uint) error
	DeleteExpired(ctx context.Context) error
}

type telegramTokenRepository struct {
	db *gorm.DB
}

func NewTelegramTokenRepository(db *gorm.DB) TelegramTokenRepository {
	return &telegramTokenRepository{db: db}
}

func (r *telegramTokenRepository) Create(ctx context.Context, token *domain.TelegramLoginToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *telegramTokenRepository) FindByToken(ctx context.Context, token string) (*domain.TelegramLoginToken, error) {
	var t domain.TelegramLoginToken
	if err := r.db.WithContext(ctx).
		Where("token = ? AND used = false AND expires_at > ?", token, time.Now()).
		First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *telegramTokenRepository) MarkAsUsed(ctx context.Context, tokenID uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.TelegramLoginToken{}).
		Where("id = ?", tokenID).
		Update("used", true).Error
}

func (r *telegramTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? OR used = true", time.Now().Add(-24*time.Hour)).
		Delete(&domain.TelegramLoginToken{}).Error
}
