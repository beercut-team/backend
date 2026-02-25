package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
)

type SyncService interface {
	Push(ctx context.Context, userID uint, req domain.SyncPushRequest) error
	Pull(ctx context.Context, userID uint, since string) (*domain.SyncPullResponse, error)
}

type syncService struct {
	repo repository.SyncRepository
}

func NewSyncService(repo repository.SyncRepository) SyncService {
	return &syncService{repo: repo}
}

func (s *syncService) Push(ctx context.Context, userID uint, req domain.SyncPushRequest) error {
	for _, m := range req.Mutations {
		clientTime, err := time.Parse(time.RFC3339, m.ClientTime)
		if err != nil {
			return errors.New("invalid client_time format, use ISO 8601")
		}

		payload, _ := json.Marshal(m.Payload)

		entry := &domain.SyncQueue{
			UserID:     userID,
			Entity:     m.Entity,
			EntityID:   m.EntityID,
			Action:     m.Action,
			Payload:    string(payload),
			ClientTime: clientTime,
		}

		if err := s.repo.Create(ctx, entry); err != nil {
			return errors.New("failed to push sync entry")
		}
	}
	return nil
}

func (s *syncService) Pull(ctx context.Context, userID uint, since string) (*domain.SyncPullResponse, error) {
	sinceTime, err := time.Parse(time.RFC3339, since)
	if err != nil {
		return nil, errors.New("invalid since format, use ISO 8601")
	}

	changes, err := s.repo.FindChangesSince(ctx, userID, sinceTime)
	if err != nil {
		return nil, err
	}

	return &domain.SyncPullResponse{
		Changes: changes,
		Since:   since,
	}, nil
}
