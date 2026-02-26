package service

import (
	"context"
	"errors"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"github.com/beercut-team/backend-boilerplate/pkg/telegram"
	"gorm.io/gorm"
)

type PatientService interface {
	Create(ctx context.Context, req domain.CreatePatientRequest, doctorID uint) (*domain.Patient, error)
	GetByID(ctx context.Context, id uint) (*domain.Patient, error)
	GetByAccessCode(ctx context.Context, code string) (*domain.PatientPublicResponse, error)
	List(ctx context.Context, filters repository.PatientFilters, offset, limit int) ([]domain.Patient, int64, error)
	Update(ctx context.Context, id uint, req domain.UpdatePatientRequest) (*domain.Patient, error)
	Delete(ctx context.Context, id uint) error
	ChangeStatus(ctx context.Context, id uint, req domain.PatientStatusRequest, changedBy uint) error
	RegenerateAccessCode(ctx context.Context, id uint) (*domain.Patient, error)
	DashboardStats(ctx context.Context, doctorID *uint) (map[domain.PatientStatus]int64, error)
}

type patientService struct {
	repo          repository.PatientRepository
	checklistRepo repository.ChecklistRepository
	notifRepo     repository.NotificationRepository
	bot           *telegram.Bot
}

func NewPatientService(repo repository.PatientRepository, checklistRepo repository.ChecklistRepository, notifRepo repository.NotificationRepository, bot *telegram.Bot) PatientService {
	return &patientService{repo: repo, checklistRepo: checklistRepo, notifRepo: notifRepo, bot: bot}
}

func (s *patientService) Create(ctx context.Context, req domain.CreatePatientRequest, doctorID uint) (*domain.Patient, error) {
	var dob time.Time
	if req.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return nil, errors.New("неверный формат даты рождения, используйте ГГГГ-ММ-ДД")
		}
		dob = parsed
	}

	patient := &domain.Patient{
		AccessCode:     domain.GenerateAccessCode(),
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		MiddleName:     req.MiddleName,
		DateOfBirth:    dob,
		Phone:          req.Phone,
		Email:          req.Email,
		Address:        req.Address,
		SNILs:          req.SNILs,
		PassportSeries: req.PassportSeries,
		PassportNumber: req.PassportNumber,
		PolicyNumber:   req.PolicyNumber,
		Diagnosis:      req.Diagnosis,
		OperationType:  req.OperationType,
		Eye:            req.Eye,
		Status:         domain.PatientStatusNew,
		DoctorID:       doctorID,
		DistrictID:     req.DistrictID,
		Notes:          req.Notes,
	}

	if err := s.repo.Create(ctx, patient); err != nil {
		return nil, errors.New("не удалось создать пациента")
	}

	// Auto-generate checklist
	s.generateChecklist(ctx, patient)

	// Transition to PREPARATION
	s.repo.UpdateStatus(ctx, patient.ID, domain.PatientStatusPreparation)
	patient.Status = domain.PatientStatusPreparation
	s.repo.CreateStatusHistory(ctx, &domain.PatientStatusHistory{
		PatientID:  patient.ID,
		FromStatus: domain.PatientStatusNew,
		ToStatus:   domain.PatientStatusPreparation,
		ChangedBy:  doctorID,
		Comment:    "Пациент создан, чек-лист сгенерирован",
	})

	// Уведомить врача о новом пациенте
	if s.bot != nil {
		patientName := patient.FirstName + " " + patient.LastName
		s.bot.NotifyDoctorNewPatient(ctx, doctorID, patientName)
	}

	return patient, nil
}

func (s *patientService) generateChecklist(ctx context.Context, patient *domain.Patient) {
	templates := domain.GetChecklistTemplates(patient.OperationType)
	now := time.Now()

	var items []domain.ChecklistItem
	for _, t := range templates {
		item := domain.ChecklistItem{
			PatientID:   patient.ID,
			Name:        t.Name,
			Description: t.Description,
			Category:    t.Category,
			IsRequired:  t.IsRequired,
			Status:      domain.ChecklistStatusPending,
		}
		if t.ExpiresInDays > 0 {
			exp := now.AddDate(0, 0, t.ExpiresInDays)
			item.ExpiresAt = &exp
		}
		items = append(items, item)
	}

	if len(items) > 0 {
		s.checklistRepo.CreateItems(ctx, items)
	}
}

func (s *patientService) GetByID(ctx context.Context, id uint) (*domain.Patient, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пациент не найден")
		}
		return nil, err
	}
	return p, nil
}

func (s *patientService) GetByAccessCode(ctx context.Context, code string) (*domain.PatientPublicResponse, error) {
	p, err := s.repo.FindByAccessCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пациент не найден")
		}
		return nil, err
	}

	history, _ := s.repo.FindStatusHistory(ctx, p.ID)

	return &domain.PatientPublicResponse{
		AccessCode:    p.AccessCode,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		Status:        p.Status,
		SurgeryDate:   p.SurgeryDate,
		StatusHistory: history,
	}, nil
}

func (s *patientService) List(ctx context.Context, filters repository.PatientFilters, offset, limit int) ([]domain.Patient, int64, error) {
	return s.repo.FindAll(ctx, filters, offset, limit)
}

func (s *patientService) Update(ctx context.Context, id uint, req domain.UpdatePatientRequest) (*domain.Patient, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пациент не найден")
		}
		return nil, err
	}

	if req.FirstName != nil {
		p.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		p.LastName = *req.LastName
	}
	if req.MiddleName != nil {
		p.MiddleName = *req.MiddleName
	}
	if req.Phone != nil {
		p.Phone = *req.Phone
	}
	if req.Email != nil {
		p.Email = *req.Email
	}
	if req.Address != nil {
		p.Address = *req.Address
	}
	if req.Diagnosis != nil {
		p.Diagnosis = *req.Diagnosis
	}
	if req.Notes != nil {
		p.Notes = *req.Notes
	}
	if req.SNILs != nil {
		p.SNILs = *req.SNILs
	}
	if req.PassportSeries != nil {
		p.PassportSeries = *req.PassportSeries
	}
	if req.PassportNumber != nil {
		p.PassportNumber = *req.PassportNumber
	}
	if req.PolicyNumber != nil {
		p.PolicyNumber = *req.PolicyNumber
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, errors.New("не удалось обновить данные пациента")
	}
	return p, nil
}

func (s *patientService) Delete(ctx context.Context, id uint) error {
	// Проверяем существование пациента
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("пациент не найден")
		}
		return err
	}

	// Удаляем пациента (каскадное удаление связанных данных настроено в GORM)
	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.New("не удалось удалить пациента")
	}

	return nil
}

func (s *patientService) ChangeStatus(ctx context.Context, id uint, req domain.PatientStatusRequest, changedBy uint) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("пациент не найден")
	}

	oldStatus := p.Status
	if err := s.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		return errors.New("не удалось обновить статус")
	}

	s.repo.CreateStatusHistory(ctx, &domain.PatientStatusHistory{
		PatientID:  id,
		FromStatus: oldStatus,
		ToStatus:   req.Status,
		ChangedBy:  changedBy,
		Comment:    req.Comment,
	})

	// Создать уведомление для пациента о смене статуса
	if s.notifRepo != nil {
		statusText := map[domain.PatientStatus]string{
			domain.PatientStatusNew:              "Новый пациент",
			domain.PatientStatusPreparation:      "На подготовке",
			domain.PatientStatusReviewNeeded:     "Отправлено на проверку хирургу",
			domain.PatientStatusApproved:         "Готов к операции",
			domain.PatientStatusSurgeryScheduled: "Операция запланирована",
			domain.PatientStatusCompleted:        "Операция завершена",
			domain.PatientStatusRejected:         "Требуется дополнительная подготовка",
		}[req.Status]

		s.notifRepo.Create(ctx, &domain.Notification{
			UserID:     id, // ID пациента, не врача!
			Type:       domain.NotifStatusChange,
			Title:      "Статус изменен",
			Body:       "Ваш статус изменен на: " + statusText,
			EntityType: "patient",
			EntityID:   id,
		})
	}

	// Отправить уведомление пациенту через Telegram
	if s.bot != nil {
		s.bot.NotifyPatientStatusChange(ctx, id, string(req.Status))

		// Если статус изменился на REVIEW_NEEDED, уведомить хирургов
		if req.Status == domain.PatientStatusReviewNeeded {
			s.bot.NotifySurgeonReviewNeeded(ctx, id)
		}
	}

	return nil
}

func (s *patientService) RegenerateAccessCode(ctx context.Context, id uint) (*domain.Patient, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пациент не найден")
		}
		return nil, err
	}

	// Генерируем новый уникальный код
	var exists int64
	var newCode string
	for {
		newCode = domain.GenerateAccessCode()
		// Проверяем уникальность
		if err := s.repo.CountByAccessCode(ctx, newCode, &exists); err != nil {
			return nil, errors.New("не удалось проверить уникальность кода")
		}
		if exists == 0 {
			break
		}
	}

	p.AccessCode = newCode
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, errors.New("не удалось обновить код доступа")
	}

	// Уведомить пациента о новом коде через Telegram
	if s.bot != nil {
		s.bot.NotifyPatientNewAccessCode(ctx, id, newCode)
	}

	return p, nil
}

func (s *patientService) DashboardStats(ctx context.Context, doctorID *uint) (map[domain.PatientStatus]int64, error) {
	return s.repo.CountByStatus(ctx, doctorID)
}
