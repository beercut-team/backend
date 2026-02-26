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
	BatchUpdate(ctx context.Context, id uint, req domain.BatchUpdateRequest, userID uint) (*domain.BatchUpdateResponse, error)
}

type patientService struct {
	db            *gorm.DB
	repo          repository.PatientRepository
	checklistRepo repository.ChecklistRepository
	notifRepo     repository.NotificationRepository
	bot           *telegram.Bot
}

func NewPatientService(db *gorm.DB, repo repository.PatientRepository, checklistRepo repository.ChecklistRepository, notifRepo repository.NotificationRepository, bot *telegram.Bot) PatientService {
	return &patientService{db: db, repo: repo, checklistRepo: checklistRepo, notifRepo: notifRepo, bot: bot}
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
		Status:         domain.PatientStatusDraft,
		DoctorID:       doctorID,
		DistrictID:     req.DistrictID,
		Notes:          req.Notes,
	}

	if err := s.repo.Create(ctx, patient); err != nil {
		return nil, errors.New("не удалось создать пациента")
	}

	// Auto-generate checklist
	s.generateChecklist(ctx, patient)

	// Transition to IN_PROGRESS
	s.repo.UpdateStatus(ctx, patient.ID, domain.PatientStatusInProgress)
	patient.Status = domain.PatientStatusInProgress
	s.repo.CreateStatusHistory(ctx, &domain.PatientStatusHistory{
		PatientID:  patient.ID,
		FromStatus: domain.PatientStatusDraft,
		ToStatus:   domain.PatientStatusInProgress,
		ChangedBy:  doctorID,
		Comment:    "Пациент создан, чек-лист сгенерирован",
	})

	// Уведомить врача о новом пациенте
	if s.bot != nil {
		patientName := patient.FirstName + " " + patient.LastName
		s.bot.NotifyDoctorNewPatient(ctx, doctorID, patientName)
		log.Info().Uint("doctor_id", doctorID).Uint("patient_id", patient.ID).Msg("попытка уведомления врача о новом пациенте")
	} else {
		log.Warn().Uint("doctor_id", doctorID).Uint("patient_id", patient.ID).Msg("Telegram бот не настроен, уведомление врачу не отправлено")
	}

	patient.PopulateDisplayNames()
	return patient, nil
}

func (s *patientService) generateChecklist(ctx context.Context, patient *domain.Patient) {
	templates := domain.GetChecklistTemplates(patient.OperationType)
	now := time.Now()

	log.Info().Uint("patient_id", patient.ID).Str("operation_type", string(patient.OperationType)).Int("templates_count", len(templates)).Msg("🔧 НАЧАЛО генерации чек-листа")

	var items []domain.ChecklistItem
	for i, t := range templates {
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
		log.Info().Int("index", i).Str("name", t.Name).Bool("template_required", t.IsRequired).Bool("item_required", item.IsRequired).Msg("📝 создание пункта")
		items = append(items, item)
	}

	if len(items) > 0 {
		if err := s.checklistRepo.CreateItems(ctx, items); err != nil {
			log.Error().Err(err).Uint("patient_id", patient.ID).Int("items_count", len(items)).Msg("не удалось создать пункты чек-листа")
		} else {
			log.Info().Uint("patient_id", patient.ID).Int("items_count", len(items)).Msg("чек-лист успешно создан")
		}
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
	p.PopulateDisplayNames()
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
		StatusDisplay: domain.GetStatusDisplayName(p.Status),
		SurgeryDate:   p.SurgeryDate,
		StatusHistory: history,
	}, nil
}

func (s *patientService) List(ctx context.Context, filters repository.PatientFilters, offset, limit int) ([]domain.Patient, int64, error) {
	patients, total, err := s.repo.FindAll(ctx, filters, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// Populate display names for all patients
	for i := range patients {
		patients[i].PopulateDisplayNames()
	}

	return patients, total, nil
}

func (s *patientService) Update(ctx context.Context, id uint, req domain.UpdatePatientRequest) (*domain.Patient, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пациент не найден")
		}
		return nil, err
	}

	// Track changes for notifications
	diagnosisChanged := false
	oldDiagnosis := p.Diagnosis

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
	if req.Diagnosis != nil && *req.Diagnosis != oldDiagnosis {
		p.Diagnosis = *req.Diagnosis
		diagnosisChanged = true
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

	// Создать уведомление врачу при изменении диагноза
	if diagnosisChanged && s.notifRepo != nil && p.Diagnosis != "" {
		patientName := p.LastName + " " + p.FirstName
		s.notifRepo.Create(ctx, &domain.Notification{
			UserID:     p.DoctorID,
			Type:       domain.NotifStatusChange,
			Title:      "Диагноз установлен",
			Body:       fmt.Sprintf("Пациент %s: установлен диагноз - %s", patientName, p.Diagnosis),
			EntityType: "patient",
			EntityID:   id,
		})
	}

	p.PopulateDisplayNames()
	return p, nil
}

func (s *patientService) Delete(ctx context.Context, id uint) error {
	log.Info().Uint("patient_id", id).Msg("удаление пациента")

	// Проверяем существование пациента
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn().Uint("patient_id", id).Msg("попытка удалить несуществующего пациента")
			return errors.New("пациент не найден")
		}
		return err
	}

	// Удаляем пациента (каскадное удаление связанных данных настроено в GORM)
	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error().Err(err).Uint("patient_id", id).Msg("ошибка удаления пациента")
		return errors.New("не удалось удалить пациента")
	}

	log.Info().Uint("patient_id", id).Msg("пациент успешно удалён")
	return nil
}

func (s *patientService) ChangeStatus(ctx context.Context, id uint, req domain.PatientStatusRequest, changedBy uint) error {
	log.Info().Uint("patient_id", id).Str("new_status", string(req.Status)).Uint("changed_by", changedBy).Msg("смена статуса пациента")

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint("patient_id", id).Msg("пациент не найден при смене статуса")
		return errors.New("пациент не найден")
	}

	oldStatus := p.Status

	// Валидация перехода
	if !domain.ValidateStatusTransition(oldStatus, req.Status) {
		log.Warn().Str("from", string(oldStatus)).Str("to", string(req.Status)).Uint("patient_id", id).Msg("недопустимый переход статуса")
		fromName := domain.GetStatusDisplayName(oldStatus)
		toName := domain.GetStatusDisplayName(req.Status)
		return fmt.Errorf("невозможно изменить статус с '%s' на '%s'. Проверьте допустимые переходы статусов", fromName, toName)
	}

	if err := s.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		log.Error().Err(err).Uint("patient_id", id).Msg("ошибка обновления статуса")
		return errors.New("не удалось обновить статус")
	}

	s.repo.CreateStatusHistory(ctx, &domain.PatientStatusHistory{
		PatientID:  id,
		FromStatus: oldStatus,
		ToStatus:   req.Status,
		ChangedBy:  changedBy,
		Comment:    req.Comment,
	})

	// Создать уведомления для врачей о смене статуса
	if s.notifRepo != nil {
		statusText := domain.GetStatusDisplayName(req.Status)
		patientName := p.LastName + " " + p.FirstName

		// Уведомить лечащего врача
		s.notifRepo.Create(ctx, &domain.Notification{
			UserID:     p.DoctorID,
			Type:       domain.NotifStatusChange,
			Title:      "Статус пациента изменен",
			Body:       fmt.Sprintf("Пациент %s: статус изменен на %s", patientName, statusText),
			EntityType: "patient",
			EntityID:   id,
		})

		// Уведомить хирурга, если назначен
		if p.SurgeonID != nil && *p.SurgeonID != changedBy {
			s.notifRepo.Create(ctx, &domain.Notification{
				UserID:     *p.SurgeonID,
				Type:       domain.NotifStatusChange,
				Title:      "Статус пациента изменен",
				Body:       fmt.Sprintf("Пациент %s: статус изменен на %s", patientName, statusText),
				EntityType: "patient",
				EntityID:   id,
			})
		}
	}

	// Отправить уведомление пациенту через Telegram
	if s.bot != nil {
		s.bot.NotifyPatientStatusChange(ctx, id, string(req.Status))
		log.Info().Uint("patient_id", id).Str("status", string(req.Status)).Msg("попытка уведомления пациента об изменении статуса")

		// Если статус изменился на PENDING_REVIEW, уведомить хирургов
		if req.Status == domain.PatientStatusPendingReview {
			s.bot.NotifySurgeonReviewNeeded(ctx, id)
			log.Info().Uint("patient_id", id).Msg("попытка уведомления хирургов о необходимости проверки")
		}
	} else {
		log.Warn().Uint("patient_id", id).Str("status", string(req.Status)).Msg("Telegram бот не настроен, уведомления не отправлены")
	}

	log.Info().Uint("patient_id", id).Str("from", string(oldStatus)).Str("to", string(req.Status)).Msg("статус успешно изменён")
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
		log.Info().Uint("patient_id", id).Str("new_code", newCode).Msg("попытка уведомления пациента о новом коде доступа")
	} else {
		log.Warn().Uint("patient_id", id).Msg("Telegram бот не настроен, уведомление о новом коде не отправлено")
	}

	p.PopulateDisplayNames()
	return p, nil
}

func (s *patientService) DashboardStats(ctx context.Context, doctorID *uint) (map[domain.PatientStatus]int64, error) {
	return s.repo.CountByStatus(ctx, doctorID)
}

func (s *patientService) BatchUpdate(ctx context.Context, id uint, req domain.BatchUpdateRequest, userID uint) (*domain.BatchUpdateResponse, error) {
	log.Info().Uint("patient_id", id).Uint("user_id", userID).Msg("начало batch update")

	response := &domain.BatchUpdateResponse{
		Success:   true,
		Conflicts: []string{},
	}

	// Начинаем транзакцию для атомарности
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Получаем пациента
		patient, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return errors.New("пациент не найден")
		}

		// Проверяем timestamp для обнаружения конфликтов
		var clientTime time.Time
		if req.Timestamp != "" {
			clientTime, err = time.Parse(time.RFC3339, req.Timestamp)
			if err == nil && patient.UpdatedAt.After(clientTime) {
				response.Conflicts = append(response.Conflicts, "Данные пациента были изменены на сервере после вашего последнего обновления")
			}
		}

		// 1. Обновляем данные пациента
		if req.Patient != nil {
			if req.Patient.Diagnosis != nil {
				patient.Diagnosis = *req.Patient.Diagnosis
			}
			if req.Patient.Notes != nil {
				patient.Notes = *req.Patient.Notes
			}
			if req.Patient.Phone != nil {
				patient.Phone = *req.Patient.Phone
			}
			if req.Patient.Email != nil {
				patient.Email = *req.Patient.Email
			}
			if req.Patient.Address != nil {
				patient.Address = *req.Patient.Address
			}
			if req.Patient.SNILs != nil {
				patient.SNILs = *req.Patient.SNILs
			}
			if req.Patient.PassportSeries != nil {
				patient.PassportSeries = *req.Patient.PassportSeries
			}
			if req.Patient.PassportNumber != nil {
				patient.PassportNumber = *req.Patient.PassportNumber
			}
			if req.Patient.PolicyNumber != nil {
				patient.PolicyNumber = *req.Patient.PolicyNumber
			}

			if err := tx.Save(patient).Error; err != nil {
				return errors.New("не удалось обновить данные пациента: " + err.Error())
			}
			response.UpdatedItems++
		}

		// 2. Меняем статус
		if req.Status != nil {
			oldStatus := patient.Status

			// Валидация перехода
			if !domain.ValidateStatusTransition(oldStatus, req.Status.Status) {
				return errors.New("недопустимый переход статуса: " + string(oldStatus) + " → " + string(req.Status.Status))
			}

			if err := tx.Model(&domain.Patient{}).Where("id = ?", id).Update("status", req.Status.Status).Error; err != nil {
				return errors.New("не удалось обновить статус: " + err.Error())
			}

			// Создаём историю статуса
			history := &domain.PatientStatusHistory{
				PatientID:  id,
				FromStatus: oldStatus,
				ToStatus:   req.Status.Status,
				ChangedBy:  userID,
				Comment:    req.Status.Comment,
			}
			if err := tx.Create(history).Error; err != nil {
				return errors.New("не удалось создать историю статуса: " + err.Error())
			}

			patient.Status = req.Status.Status
			response.UpdatedItems++
		}

		// 3. Обновляем чек-лист
		if len(req.Checklist) > 0 {
			for _, itemUpdate := range req.Checklist {
				var item domain.ChecklistItem
				if err := tx.First(&item, itemUpdate.ID).Error; err != nil {
					response.Conflicts = append(response.Conflicts, "Элемент чек-листа не найден")
					continue
				}

				// Проверяем, что элемент принадлежит этому пациенту
				if item.PatientID != id {
					response.Conflicts = append(response.Conflicts, "Элемент чек-листа не принадлежит данному пациенту")
					continue
				}

				// Применяем обновления
				updated := false
				if itemUpdate.Status != nil {
					item.Status = domain.ChecklistItemStatus(*itemUpdate.Status)
					if item.Status == domain.ChecklistStatusCompleted {
						now := time.Now()
						item.CompletedAt = &now
						item.CompletedBy = &userID
					}
					updated = true
				}
				if itemUpdate.Result != nil {
					item.Result = *itemUpdate.Result
					updated = true
				}
				if itemUpdate.Notes != nil {
					item.Notes = *itemUpdate.Notes
					updated = true
				}

				if updated {
					if err := tx.Save(&item).Error; err != nil {
						response.Conflicts = append(response.Conflicts, "Ошибка обновления элемента чек-листа")
					} else {
						response.UpdatedItems++
					}
				}
			}

			// Проверяем автопереход статуса после обновления чек-листа
			var total, required, requiredCompleted int64
			tx.Model(&domain.ChecklistItem{}).Where("patient_id = ?", id).Count(&total)
			tx.Model(&domain.ChecklistItem{}).Where("patient_id = ? AND is_required = ?", id, true).Count(&required)
			tx.Model(&domain.ChecklistItem{}).Where("patient_id = ? AND is_required = ? AND status = ?", id, true, domain.ChecklistStatusCompleted).Count(&requiredCompleted)

			if required > 0 && required == requiredCompleted && patient.Status == domain.PatientStatusInProgress {
				if err := tx.Model(&domain.Patient{}).Where("id = ?", id).Update("status", domain.PatientStatusPendingReview).Error; err != nil {
					return errors.New("не удалось выполнить автопереход статуса")
				}

				history := &domain.PatientStatusHistory{
					PatientID:  id,
					FromStatus: domain.PatientStatusInProgress,
					ToStatus:   domain.PatientStatusPendingReview,
					Comment:    "Все обязательные пункты чек-листа выполнены (batch update)",
				}
				if err := tx.Create(history).Error; err != nil {
					return errors.New("не удалось создать историю автоперехода")
				}

				patient.Status = domain.PatientStatusPendingReview
			}
		}

		// Перезагружаем пациента для актуальных данных
		if err := tx.Preload("Doctor").Preload("Surgeon").Preload("District").First(patient, id).Error; err != nil {
			return errors.New("не удалось перезагрузить данные пациента")
		}
		response.Patient = patient

		return nil
	})

	if err != nil {
		log.Error().Err(err).Uint("patient_id", id).Msg("ошибка batch update")
		response.Success = false
		response.Message = "Пакетное обновление не выполнено: " + err.Error()
		return response, err
	}

	// Уведомления отправляем после успешной транзакции
	if req.Status != nil && s.notifRepo != nil && response.Patient != nil {
		statusText := domain.GetStatusDisplayName(req.Status.Status)
		patientName := response.Patient.LastName + " " + response.Patient.FirstName

		// Уведомить лечащего врача
		s.notifRepo.Create(ctx, &domain.Notification{
			UserID:     response.Patient.DoctorID,
			Type:       domain.NotifStatusChange,
			Title:      "Статус пациента изменен",
			Body:       fmt.Sprintf("Пациент %s: статус изменен на %s", patientName, statusText),
			EntityType: "patient",
			EntityID:   id,
		})

		// Уведомить хирурга, если назначен
		if response.Patient.SurgeonID != nil {
			s.notifRepo.Create(ctx, &domain.Notification{
				UserID:     *response.Patient.SurgeonID,
				Type:       domain.NotifStatusChange,
				Title:      "Статус пациента изменен",
				Body:       fmt.Sprintf("Пациент %s: статус изменен на %s", patientName, statusText),
				EntityType: "patient",
				EntityID:   id,
			})
		}
	}

	if req.Status != nil {
		if s.bot != nil {
			s.bot.NotifyPatientStatusChange(ctx, id, string(req.Status.Status))
			log.Info().Uint("patient_id", id).Str("status", string(req.Status.Status)).Msg("попытка уведомления пациента об изменении статуса (batch)")
			if req.Status.Status == domain.PatientStatusPendingReview {
				s.bot.NotifySurgeonReviewNeeded(ctx, id)
				log.Info().Uint("patient_id", id).Msg("попытка уведомления хирургов о необходимости проверки (batch)")
			}
		} else {
			log.Warn().Uint("patient_id", id).Str("status", string(req.Status.Status)).Msg("Telegram бот не настроен, уведомления не отправлены (batch)")
		}
	}

	// Populate display names for response
	if response.Patient != nil {
		response.Patient.PopulateDisplayNames()
	}

	log.Info().Uint("patient_id", id).Int("updated_items", response.UpdatedItems).Msg("batch update завершён успешно")
	response.Message = "Пакетное обновление выполнено успешно"
	return response, nil
}
