package service

import (
	"context"
	"errors"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"gorm.io/gorm"
)

type ChecklistService interface {
	GetByPatient(ctx context.Context, patientID uint) ([]domain.ChecklistItem, error)
	UpdateItem(ctx context.Context, id uint, req domain.UpdateChecklistItemRequest, userID uint) (*domain.ChecklistItem, error)
	ReviewItem(ctx context.Context, id uint, req domain.ReviewChecklistItemRequest, reviewerID uint) (*domain.ChecklistItem, error)
	GetProgress(ctx context.Context, patientID uint) (*ChecklistProgress, error)
	CheckAndTransition(ctx context.Context, patientID uint) error
}

type ChecklistProgress struct {
	Total             int64   `json:"total"`
	Completed         int64   `json:"completed"`
	Required          int64   `json:"required"`
	RequiredCompleted int64   `json:"required_completed"`
	Percentage        float64 `json:"percentage"`
}

type checklistService struct {
	repo        repository.ChecklistRepository
	patientRepo repository.PatientRepository
}

func NewChecklistService(repo repository.ChecklistRepository, patientRepo repository.PatientRepository) ChecklistService {
	return &checklistService{repo: repo, patientRepo: patientRepo}
}

func (s *checklistService) GetByPatient(ctx context.Context, patientID uint) ([]domain.ChecklistItem, error) {
	return s.repo.FindItemsByPatient(ctx, patientID)
}

func (s *checklistService) UpdateItem(ctx context.Context, id uint, req domain.UpdateChecklistItemRequest, userID uint) (*domain.ChecklistItem, error) {
	item, err := s.repo.FindItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("checklist item not found")
		}
		return nil, err
	}

	if req.Status != "" {
		status := domain.ChecklistItemStatus(req.Status)
		item.Status = status
		if status == domain.ChecklistStatusCompleted {
			now := time.Now()
			item.CompletedAt = &now
			item.CompletedBy = &userID
		}
	}
	if req.Result != nil {
		item.Result = *req.Result
	}
	if req.Notes != nil {
		item.Notes = *req.Notes
	}

	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return nil, errors.New("failed to update checklist item")
	}

	// Check if all required items are completed
	s.CheckAndTransition(ctx, item.PatientID)

	return item, nil
}

func (s *checklistService) ReviewItem(ctx context.Context, id uint, req domain.ReviewChecklistItemRequest, reviewerID uint) (*domain.ChecklistItem, error) {
	item, err := s.repo.FindItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("checklist item not found")
		}
		return nil, err
	}

	status := domain.ChecklistItemStatus(req.Status)
	if status != domain.ChecklistStatusCompleted && status != domain.ChecklistStatusRejected {
		return nil, errors.New("review status must be COMPLETED or REJECTED")
	}

	item.Status = status
	item.ReviewedBy = &reviewerID
	item.ReviewNote = req.ReviewNote

	if status == domain.ChecklistStatusCompleted {
		now := time.Now()
		item.CompletedAt = &now
	}

	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return nil, errors.New("failed to review checklist item")
	}

	s.CheckAndTransition(ctx, item.PatientID)
	return item, nil
}

func (s *checklistService) GetProgress(ctx context.Context, patientID uint) (*ChecklistProgress, error) {
	total, completed, required, requiredCompleted, err := s.repo.CountByPatient(ctx, patientID)
	if err != nil {
		return nil, err
	}

	var pct float64
	if total > 0 {
		pct = float64(completed) / float64(total) * 100
	}

	return &ChecklistProgress{
		Total:             total,
		Completed:         completed,
		Required:          required,
		RequiredCompleted: requiredCompleted,
		Percentage:        pct,
	}, nil
}

func (s *checklistService) CheckAndTransition(ctx context.Context, patientID uint) error {
	_, _, required, requiredCompleted, err := s.repo.CountByPatient(ctx, patientID)
	if err != nil {
		return err
	}

	if required > 0 && required == requiredCompleted {
		p, err := s.patientRepo.FindByID(ctx, patientID)
		if err != nil {
			return err
		}
		if p.Status == domain.PatientStatusPreparation {
			s.patientRepo.UpdateStatus(ctx, patientID, domain.PatientStatusReviewNeeded)
			s.patientRepo.CreateStatusHistory(ctx, &domain.PatientStatusHistory{
				PatientID:  patientID,
				FromStatus: domain.PatientStatusPreparation,
				ToStatus:   domain.PatientStatusReviewNeeded,
				Comment:    "All required checklist items completed",
			})
		}
	}
	return nil
}
