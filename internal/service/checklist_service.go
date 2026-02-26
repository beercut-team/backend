package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"github.com/beercut-team/backend-boilerplate/pkg/telegram"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ChecklistService interface {
	GetByPatient(ctx context.Context, patientID uint) ([]domain.ChecklistItem, error)
	CreateItem(ctx context.Context, req domain.CreateChecklistItemRequest, userID uint) (*domain.ChecklistItem, error)
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
	notifRepo   repository.NotificationRepository
	bot         *telegram.Bot
}

func NewChecklistService(repo repository.ChecklistRepository, patientRepo repository.PatientRepository, notifRepo repository.NotificationRepository, bot *telegram.Bot) ChecklistService {
	return &checklistService{
		repo:        repo,
		patientRepo: patientRepo,
		notifRepo:   notifRepo,
		bot:         bot,
	}
}

func (s *checklistService) GetByPatient(ctx context.Context, patientID uint) ([]domain.ChecklistItem, error) {
	return s.repo.FindItemsByPatient(ctx, patientID)
}

func (s *checklistService) CreateItem(ctx context.Context, req domain.CreateChecklistItemRequest, userID uint) (*domain.ChecklistItem, error) {
	// Verify patient exists
	patient, err := s.patientRepo.FindByID(ctx, req.PatientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пациент не найден")
		}
		return nil, err
	}

	item := &domain.ChecklistItem{
		PatientID:   req.PatientID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		IsRequired:  req.IsRequired,
		Status:      domain.ChecklistStatusPending,
	}

	// Set expiration if provided
	if req.ExpiresInDays > 0 {
		exp := time.Now().AddDate(0, 0, req.ExpiresInDays)
		item.ExpiresAt = &exp
	}

	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, errors.New("не удалось создать пункт чек-листа")
	}

	// Создать уведомление в БД для врача
	if s.notifRepo != nil {
		patientName := patient.LastName + " " + patient.FirstName
		notifBody := fmt.Sprintf("Пациент %s: добавлен пункт чек-листа \"%s\"", patientName, item.Name)
		if item.IsRequired {
			notifBody += " (обязательный)"
		}

		s.notifRepo.Create(ctx, &domain.Notification{
			UserID:     patient.DoctorID,
			Type:       domain.NotifStatusChange,
			Title:      "Новый пункт чек-листа",
			Body:       notifBody,
			EntityType: "checklist_item",
			EntityID:   item.ID,
		})

		// Уведомить хирурга, если назначен
		if patient.SurgeonID != nil {
			s.notifRepo.Create(ctx, &domain.Notification{
				UserID:     *patient.SurgeonID,
				Type:       domain.NotifStatusChange,
				Title:      "Новый пункт чек-листа",
				Body:       notifBody,
				EntityType: "checklist_item",
				EntityID:   item.ID,
			})
		}
	}

	// Отправить уведомление пациенту через Telegram
	if s.bot != nil {
		message := fmt.Sprintf("📋 Добавлен новый пункт в чек-лист\n\n%s", item.Name)
		if item.Description != "" {
			message += fmt.Sprintf("\n%s", item.Description)
		}
		if item.IsRequired {
			message += "\n\n⚠️ Обязательный пункт"
		}
		s.bot.NotifyPatient(ctx, req.PatientID, message)
		log.Info().Uint("patient_id", req.PatientID).Str("item_name", item.Name).Msg("попытка отправки уведомления о новом пункте чек-листа")
	} else {
		log.Debug().Uint("patient_id", req.PatientID).Msg("уведомление пациенту пропущено: Telegram бот не инициализирован")
	}

	return item, nil
}

func (s *checklistService) UpdateItem(ctx context.Context, id uint, req domain.UpdateChecklistItemRequest, userID uint) (*domain.ChecklistItem, error) {
	item, err := s.repo.FindItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("элемент чек-листа не найден")
		}
		return nil, err
	}

	oldStatus := item.Status
	statusChanged := false

	if req.Status != "" {
		status := domain.ChecklistItemStatus(req.Status)
		item.Status = status
		statusChanged = (oldStatus != status)
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
		return nil, errors.New("не удалось обновить элемент чек-листа")
	}

	// Создать уведомление в БД для врача при изменении статуса
	if statusChanged && s.notifRepo != nil {
		patient, err := s.patientRepo.FindByID(ctx, item.PatientID)
		if err == nil {
			patientName := patient.LastName + " " + patient.FirstName
			statusName := string(item.Status)
			notifBody := fmt.Sprintf("Пациент %s: пункт чек-листа \"%s\" изменён на %s", patientName, item.Name, statusName)
			if item.Result != "" {
				notifBody += fmt.Sprintf(" (результат: %s)", item.Result)
			}

			s.notifRepo.Create(ctx, &domain.Notification{
				UserID:     patient.DoctorID,
				Type:       domain.NotifStatusChange,
				Title:      "Обновление чек-листа",
				Body:       notifBody,
				EntityType: "checklist_item",
				EntityID:   item.ID,
			})

			// Уведомить хирурга, если назначен
			if patient.SurgeonID != nil {
				s.notifRepo.Create(ctx, &domain.Notification{
					UserID:     *patient.SurgeonID,
					Type:       domain.NotifStatusChange,
					Title:      "Обновление чек-листа",
					Body:       notifBody,
					EntityType: "checklist_item",
					EntityID:   item.ID,
				})
			}
		}
	}

	// Отправить уведомление пациенту при изменении статуса
	if statusChanged {
		if s.bot != nil {
			var message string
			switch item.Status {
			case domain.ChecklistStatusCompleted:
				message = fmt.Sprintf("✅ Пункт чек-листа выполнен\n\n%s", item.Name)
				if item.Result != "" {
					message += fmt.Sprintf("\n\nРезультат: %s", item.Result)
				}
			case domain.ChecklistStatusInProgress:
				message = fmt.Sprintf("⏳ Пункт чек-листа в работе\n\n%s", item.Name)
			case domain.ChecklistStatusRejected:
				message = fmt.Sprintf("❌ Пункт чек-листа отклонён\n\n%s", item.Name)
				if item.Notes != "" {
					message += fmt.Sprintf("\n\nПримечание: %s", item.Notes)
				}
			}
			if message != "" {
				s.bot.NotifyPatient(ctx, item.PatientID, message)
				log.Info().Uint("patient_id", item.PatientID).Str("item_name", item.Name).Str("status", string(item.Status)).Msg("попытка отправки уведомления об обновлении чек-листа")
			}
		} else {
			log.Debug().Uint("patient_id", item.PatientID).Msg("уведомление пациенту пропущено: Telegram бот не инициализирован")
		}
	}

	// Check if all required items are completed
	s.CheckAndTransition(ctx, item.PatientID)

	return item, nil
}

func (s *checklistService) ReviewItem(ctx context.Context, id uint, req domain.ReviewChecklistItemRequest, reviewerID uint) (*domain.ChecklistItem, error) {
	item, err := s.repo.FindItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("элемент чек-листа не найден")
		}
		return nil, err
	}

	status := domain.ChecklistItemStatus(req.Status)
	if status != domain.ChecklistStatusCompleted && status != domain.ChecklistStatusRejected {
		return nil, errors.New("статус проверки должен быть COMPLETED или REJECTED")
	}

	item.Status = status
	item.ReviewedBy = &reviewerID
	item.ReviewNote = req.ReviewNote

	if status == domain.ChecklistStatusCompleted {
		now := time.Now()
		item.CompletedAt = &now
	}

	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return nil, errors.New("не удалось проверить элемент чек-листа")
	}

	// Создать уведомление в БД для врача о результате проверки
	if s.notifRepo != nil {
		patient, err := s.patientRepo.FindByID(ctx, item.PatientID)
		if err == nil {
			patientName := patient.LastName + " " + patient.FirstName
			var notifTitle, notifBody string

			if status == domain.ChecklistStatusCompleted {
				notifTitle = "Пункт чек-листа одобрен"
				notifBody = fmt.Sprintf("Пациент %s: хирург одобрил пункт \"%s\"", patientName, item.Name)
			} else {
				notifTitle = "Пункт чек-листа отклонён"
				notifBody = fmt.Sprintf("Пациент %s: хирург отклонил пункт \"%s\"", patientName, item.Name)
			}

			if req.ReviewNote != "" {
				notifBody += fmt.Sprintf(" (комментарий: %s)", req.ReviewNote)
			}

			// Уведомить лечащего врача
			s.notifRepo.Create(ctx, &domain.Notification{
				UserID:     patient.DoctorID,
				Type:       domain.NotifStatusChange,
				Title:      notifTitle,
				Body:       notifBody,
				EntityType: "checklist_item",
				EntityID:   item.ID,
			})
		}
	}

	// Отправить уведомление пациенту о результате проверки
	if s.bot != nil {
		var message string
		if status == domain.ChecklistStatusCompleted {
			message = fmt.Sprintf("✅ Хирург одобрил пункт чек-листа\n\n%s", item.Name)
			if req.ReviewNote != "" {
				message += fmt.Sprintf("\n\nКомментарий: %s", req.ReviewNote)
			}
		} else if status == domain.ChecklistStatusRejected {
			message = fmt.Sprintf("❌ Хирург отклонил пункт чек-листа\n\n%s", item.Name)
			if req.ReviewNote != "" {
				message += fmt.Sprintf("\n\nПричина: %s", req.ReviewNote)
			}
			message += "\n\nОбратитесь к врачу для уточнения деталей."
		}
		if message != "" {
			s.bot.NotifyPatient(ctx, item.PatientID, message)
			log.Info().Uint("patient_id", item.PatientID).Str("item_name", item.Name).Str("review_status", string(status)).Msg("попытка отправки уведомления о проверке чек-листа")
		}
	} else {
		log.Debug().Uint("patient_id", item.PatientID).Msg("уведомление пациенту пропущено: Telegram бот не инициализирован")
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
		log.Error().Err(err).Uint("patient_id", patientID).Msg("ошибка подсчёта пунктов чек-листа")
		return err
	}

	log.Info().Uint("patient_id", patientID).Int64("required", required).Int64("required_completed", requiredCompleted).Msg("проверка автоперехода статуса")

	if required > 0 && required == requiredCompleted {
		p, err := s.patientRepo.FindByID(ctx, patientID)
		if err != nil {
			log.Error().Err(err).Uint("patient_id", patientID).Msg("не удалось найти пациента для автоперехода")
			return err
		}

		log.Info().Uint("patient_id", patientID).Str("current_status", string(p.Status)).Msg("все обязательные пункты выполнены")

		if p.Status == domain.PatientStatusInProgress {
			if err := s.patientRepo.UpdateStatus(ctx, patientID, domain.PatientStatusPendingReview); err != nil {
				log.Error().Err(err).Uint("patient_id", patientID).Msg("не удалось обновить статус пациента")
				return err
			}

			if err := s.patientRepo.CreateStatusHistory(ctx, &domain.PatientStatusHistory{
				PatientID:  patientID,
				FromStatus: domain.PatientStatusInProgress,
				ToStatus:   domain.PatientStatusPendingReview,
				Comment:    "Все обязательные пункты чек-листа выполнены",
			}); err != nil {
				log.Error().Err(err).Uint("patient_id", patientID).Msg("не удалось создать историю статуса")
			}

			log.Info().Uint("patient_id", patientID).Msg("статус автоматически изменён на PENDING_REVIEW")
		} else {
			log.Info().Uint("patient_id", patientID).Str("current_status", string(p.Status)).Msg("статус не IN_PROGRESS, автопереход не требуется")
		}
	} else {
		log.Debug().Uint("patient_id", patientID).Int64("required", required).Int64("required_completed", requiredCompleted).Msg("не все обязательные пункты выполнены")
	}
	return nil
}
