package service

import (
	"context"
	"encoding/json"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
)

type AuditService interface {
	LogAction(ctx context.Context, userID uint, action, entity string, entityID uint, oldValue, newValue interface{}, ip string) error
}

type auditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) LogAction(ctx context.Context, userID uint, action, entity string, entityID uint, oldValue, newValue interface{}, ip string) error {
	var oldJSON, newJSON string

	if oldValue != nil {
		if b, err := json.Marshal(oldValue); err == nil {
			oldJSON = string(b)
		}
	}

	if newValue != nil {
		if b, err := json.Marshal(newValue); err == nil {
			newJSON = string(b)
		}
	}

	log := &domain.AuditLog{
		UserID:   userID,
		Action:   action,
		Entity:   entity,
		EntityID: entityID,
		OldValue: oldJSON,
		NewValue: newJSON,
		IP:       ip,
	}

	return s.repo.Create(ctx, log)
}
