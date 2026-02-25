package service

import (
	"context"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
)

type NotificationService interface {
	Create(ctx context.Context, req domain.CreateNotificationRequest) (*domain.Notification, error)
	List(ctx context.Context, userID uint, offset, limit int) ([]domain.Notification, int64, error)
	MarkAsRead(ctx context.Context, id, userID uint) error
	MarkAllAsRead(ctx context.Context, userID uint) error
	UnreadCount(ctx context.Context, userID uint) (int64, error)
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) Create(ctx context.Context, req domain.CreateNotificationRequest) (*domain.Notification, error) {
	n := &domain.Notification{
		UserID:     req.UserID,
		Type:       req.Type,
		Title:      req.Title,
		Body:       req.Body,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *notificationService) List(ctx context.Context, userID uint, offset, limit int) ([]domain.Notification, int64, error) {
	return s.repo.FindByUser(ctx, userID, offset, limit)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id, userID uint) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uint) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) UnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.repo.UnreadCount(ctx, userID)
}
