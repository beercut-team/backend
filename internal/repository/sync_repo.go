package repository

import (
	"context"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type SyncRepository interface {
	Create(ctx context.Context, entry *domain.SyncQueue) error
	FindChangesSince(ctx context.Context, userID uint, since time.Time) ([]domain.SyncQueue, error)
}

type syncRepository struct {
	db *gorm.DB
}

func NewSyncRepository(db *gorm.DB) SyncRepository {
	return &syncRepository{db: db}
}

func (r *syncRepository) Create(ctx context.Context, entry *domain.SyncQueue) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *syncRepository) FindChangesSince(ctx context.Context, userID uint, since time.Time) ([]domain.SyncQueue, error) {
	var entries []domain.SyncQueue
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND server_time > ?", userID, since).
		Order("server_time ASC").Find(&entries).Error
	return entries, err
}
