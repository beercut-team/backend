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
	Delete(ctx context.Context, id uint, deletedBy uint) error
}

type surgeryService struct {
	repo          repository.SurgeryRepository
	patientRepo   repository.PatientRepository
	checklistRepo repository.ChecklistRepository
	notifRepo     repository.NotificationRepository
}

func NewSurgeryService(repo repository.SurgeryRepository, patientRepo repository.PatientRepository, checklistRepo repository.ChecklistRepository, notifRepo repository.NotificationRepository) SurgeryService {
	return &surgeryService{repo: repo, patientRepo: patientRepo, checklistRepo: checklistRepo, notifRepo: notifRepo}
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
		Comment:    "Операция запланирована на " + req.ScheduledDate,
	})

	// Update patient surgery date and surgeon
	patient.SurgeryDate = &date
	patient.SurgeonID = &surgeonID
	s.patientRepo.Update(ctx, patient)

	// Создать уведомление для пациента о запланированной операции
	if s.notifRepo != nil {
		s.notifRepo.Create(ctx, &domain.Notification{
			UserID:     surgery.PatientID, // ID пациента
			Type:       domain.NotifSurgeryScheduled,
			Title:      "Операция запланирована",
			Body:       "Ваша операция назначена на " + date.Format("02.01.2006"),
			EntityType: "surgery",
			EntityID:   surgery.ID,
		})
	}

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

func (s *surgeryService) Delete(ctx context.Context, id uint, deletedBy uint) error {
	surgery, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("операция не найдена")
		}
		return err
	}

	// Get patient to revert status
	patient, err := s.patientRepo.FindByID(ctx, surgery.PatientID)
	if err != nil {
		return errors.New("пациент не найден")
	}

	// Delete surgery
	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.New("не удалось удалить операцию")
	}

	// Revert patient status to APPROVED if surgery was scheduled
	if surgery.Status == domain.SurgeryStatusScheduled && patient.Status == domain.PatientStatusSurgeryScheduled {
		s.patientRepo.UpdateStatus(ctx, surgery.PatientID, domain.PatientStatusApproved)
		s.patientRepo.CreateStatusHistory(ctx, &domain.PatientStatusHistory{
			PatientID:  surgery.PatientID,
			FromStatus: patient.Status,
			ToStatus:   domain.PatientStatusApproved,
			ChangedBy:  deletedBy,
			Comment:    "Операция отменена",
		})

		// Clear surgery date and surgeon
		patient.SurgeryDate = nil
		patient.SurgeonID = nil
		s.patientRepo.Update(ctx, patient)
	}

	return nil
}
