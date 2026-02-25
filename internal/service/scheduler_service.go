package service

import (
	"context"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type SchedulerService struct {
	cron            *cron.Cron
	checklistRepo   repository.ChecklistRepository
	surgeryRepo     repository.SurgeryRepository
	notifRepo       repository.NotificationRepository
	mediaRepo       repository.MediaRepository
}

func NewSchedulerService(
	checklistRepo repository.ChecklistRepository,
	surgeryRepo repository.SurgeryRepository,
	notifRepo repository.NotificationRepository,
	mediaRepo repository.MediaRepository,
) *SchedulerService {
	return &SchedulerService{
		cron:          cron.New(),
		checklistRepo: checklistRepo,
		surgeryRepo:   surgeryRepo,
		notifRepo:     notifRepo,
		mediaRepo:     mediaRepo,
	}
}

func (s *SchedulerService) Start() {
	// Daily 02:00 — check checklist expiry
	s.cron.AddFunc("0 2 * * *", s.checkExpiredItems)

	// Daily 09:00 — surgery reminders
	s.cron.AddFunc("0 9 * * *", s.sendSurgeryReminders)

	// Daily 03:00 — cleanup orphaned media
	s.cron.AddFunc("0 3 * * *", s.cleanupOrphanedMedia)

	s.cron.Start()
	log.Info().Msg("планировщик запущен")
}

func (s *SchedulerService) Stop() {
	s.cron.Stop()
}

func (s *SchedulerService) checkExpiredItems() {
	ctx := context.Background()
	items, err := s.checklistRepo.FindExpiredItems(ctx)
	if err != nil {
		log.Error().Err(err).Msg("планировщик: не удалось найти просроченные пункты")
		return
	}

	for _, item := range items {
		s.checklistRepo.UpdateItemStatus(ctx, item.ID, domain.ChecklistStatusExpired)
		log.Info().Uint("item_id", item.ID).Msg("планировщик: пункт чек-листа отмечен как просроченный")
	}
}

func (s *SchedulerService) sendSurgeryReminders() {
	ctx := context.Background()

	// Remind 3 days before
	threeDays := time.Now().AddDate(0, 0, 3)
	surgeries, err := s.surgeryRepo.FindUpcoming(ctx, threeDays)
	if err != nil {
		log.Error().Err(err).Msg("планировщик: не удалось найти предстоящие операции")
		return
	}

	for _, surgery := range surgeries {
		daysUntil := int(time.Until(surgery.ScheduledDate).Hours() / 24)
		if daysUntil == 3 || daysUntil == 1 {
			s.notifRepo.Create(ctx, &domain.Notification{
				UserID:     surgery.SurgeonID,
				Type:       domain.NotifSurgeryReminder,
				Title:      "Напоминание об операции",
				Body:       surgery.Patient.LastName + " " + surgery.Patient.FirstName + " — операция через " + time.Until(surgery.ScheduledDate).Round(24*time.Hour).String(),
				EntityType: "surgery",
				EntityID:   surgery.ID,
			})
		}
	}
}

func (s *SchedulerService) cleanupOrphanedMedia() {
	ctx := context.Background()
	orphaned, err := s.mediaRepo.FindOrphaned(ctx)
	if err != nil {
		log.Error().Err(err).Msg("планировщик: не удалось найти потерянные медиафайлы")
		return
	}

	for _, m := range orphaned {
		s.mediaRepo.Delete(ctx, m.ID)
		log.Info().Uint("media_id", m.ID).Msg("планировщик: удалён потерянный медиафайл")
	}
}
