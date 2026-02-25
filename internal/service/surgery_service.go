package service

import (
	"context"
	"errors"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"gorm.io/gorm"
)

type SurgeryService interface {
	Schedule(ctx context.Context, req domain.CreateSurgeryRequest, surgeonID uint) (*domain.Surgery, error)
	GetByID(ctx context.Context, id uint) (*domain.Surgery, error)
	ListBySurgeon(ctx context.Context, surgeonID uint, offset, limit int) ([]domain.Surgery, int64, error)
	Update(ctx context.Context, id uint, req domain.UpdateSurgeryRequest) (*domain.Surgery, error)
}

type surgeryService struct {
	repo          repository.SurgeryRepository
	patientRepo   repository.PatientRepository
	checklistRepo repository.ChecklistRepository
}

func NewSurgeryService(repo repository.SurgeryRepository, patientRepo repository.PatientRepository, checklistRepo repository.ChecklistRepository) SurgeryService {
	return &surgeryService{repo: repo, patientRepo: patientRepo, checklistRepo: checklistRepo}
}

func (s *surgeryService) Schedule(ctx context.Context, req domain.CreateSurgeryRequest, surgeonID uint) (*domain.Surgery, error) {
	patient, err := s.patientRepo.FindByID(ctx, req.PatientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пациент не найден")
		}
		return nil, err
	}

	// Readiness check
	_, _, required, requiredCompleted, err := s.checklistRepo.CountByPatient(ctx, req.PatientID)
	if err != nil {
		return nil, err
	}
	if required != requiredCompleted {
		return nil, errors.New("не все обязательные пункты чек-листа выполнены")
	}

	date, err := time.Parse("2006-01-02", req.ScheduledDate)
	if err != nil {
		return nil, errors.New("неверный формат даты, используйте ГГГГ-ММ-ДД")
	}

	surgery := &domain.Surgery{
		PatientID:     req.PatientID,
		SurgeonID:     surgeonID,
		ScheduledDate: date,
		OperationType: patient.OperationType,
		Eye:           patient.Eye,
		Status:        domain.SurgeryStatusScheduled,
		Notes:         req.Notes,
	}

	if err := s.repo.Create(ctx, surgery); err != nil {
		return nil, errors.New("не удалось запланировать операцию")
	}

	// Auto-transition patient
	s.patientRepo.UpdateStatus(ctx, req.PatientID, domain.PatientStatusSurgeryScheduled)
	s.patientRepo.CreateStatusHistory(ctx, &domain.PatientStatusHistory{
		PatientID:  req.PatientID,
		FromStatus: patient.Status,
		ToStatus:   domain.PatientStatusSurgeryScheduled,
		ChangedBy:  surgeonID,
		Comment:    "Surgery scheduled for " + req.ScheduledDate,
	})

	// Update patient surgery date and surgeon
	patient.SurgeryDate = &date
	patient.SurgeonID = &surgeonID
	s.patientRepo.Update(ctx, patient)

	return surgery, nil
}

func (s *surgeryService) GetByID(ctx context.Context, id uint) (*domain.Surgery, error) {
	surgery, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("операция не найдена")
		}
		return nil, err
	}
	return surgery, nil
}

func (s *surgeryService) ListBySurgeon(ctx context.Context, surgeonID uint, offset, limit int) ([]domain.Surgery, int64, error) {
	return s.repo.FindBySurgeon(ctx, surgeonID, offset, limit)
}

func (s *surgeryService) Update(ctx context.Context, id uint, req domain.UpdateSurgeryRequest) (*domain.Surgery, error) {
	surgery, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("операция не найдена")
		}
		return nil, err
	}

	if req.ScheduledDate != nil {
		date, err := time.Parse("2006-01-02", *req.ScheduledDate)
		if err != nil {
			return nil, errors.New("неверный формат даты")
		}
		surgery.ScheduledDate = date
	}
	if req.Status != nil {
		surgery.Status = domain.SurgeryStatus(*req.Status)
	}
	if req.Notes != nil {
		surgery.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, surgery); err != nil {
		return nil, errors.New("не удалось обновить операцию")
	}
	return surgery, nil
}
