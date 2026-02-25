package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type Bot struct {
	api          *tgbotapi.BotAPI
	patientRepo  repository.PatientRepository
	telegramRepo repository.TelegramRepository
}

func NewBot(token string, patientRepo repository.PatientRepository, telegramRepo repository.TelegramRepository) (*Bot, error) {
	if token == "" {
		return nil, nil
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать Telegram бот: %w", err)
	}

	log.Info().Str("bot", api.Self.UserName).Msg("Telegram бот авторизован")

	return &Bot{
		api:          api,
		patientRepo:  patientRepo,
		telegramRepo: telegramRepo,
	}, nil
}

func (b *Bot) Start() {
	if b == nil || b.api == nil {
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}
			b.handleMessage(update.Message)
		}
	}()

	log.Info().Msg("Telegram бот слушает обновления")
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	ctx := context.Background()
	text := strings.TrimSpace(msg.Text)

	switch {
	case strings.HasPrefix(text, "/start"):
		b.handleStart(ctx, msg)
	case text == "/status":
		b.handleStatus(ctx, msg)
	case text == "/help":
		b.sendMessage(msg.Chat.ID, "Доступные команды:\n/start <код_доступа> — Привязать к вашей карте пациента\n/status — Проверить текущий статус подготовки\n/help — Показать эту справку")
	default:
		b.sendMessage(msg.Chat.ID, "Неизвестная команда. Используйте /help для просмотра доступных команд.")
	}
}

func (b *Bot) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	parts := strings.Fields(msg.Text)
	if len(parts) < 2 {
		b.sendMessage(msg.Chat.ID, "Пожалуйста, укажите ваш код доступа: /start <код_доступа>")
		return
	}

	accessCode := parts[1]
	patient, err := b.patientRepo.FindByAccessCode(ctx, accessCode)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Неверный код доступа. Пожалуйста, проверьте и попробуйте снова.")
		return
	}

	// Check if already bound
	existing, _ := b.telegramRepo.FindByChatID(ctx, msg.Chat.ID)
	if existing != nil {
		b.sendMessage(msg.Chat.ID, "Этот чат уже привязан. Используйте /status для проверки статуса.")
		return
	}

	binding := &domain.TelegramBinding{
		PatientID:  patient.ID,
		ChatID:     msg.Chat.ID,
		AccessCode: accessCode,
		IsActive:   true,
	}

	if err := b.telegramRepo.Create(ctx, binding); err != nil {
		b.sendMessage(msg.Chat.ID, "Не удалось привязать. Пожалуйста, попробуйте позже.")
		return
	}

	b.sendMessage(msg.Chat.ID, fmt.Sprintf(
		"Успешно привязано!\nПациент: %s %s\nСтатус: %s\n\nИспользуйте /status для проверки прогресса подготовки.",
		patient.FirstName, patient.LastName, patient.Status,
	))
}

func (b *Bot) handleStatus(ctx context.Context, msg *tgbotapi.Message) {
	binding, err := b.telegramRepo.FindByChatID(ctx, msg.Chat.ID)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Пациент не привязан. Сначала используйте /start <код_доступа>.")
		return
	}

	patient, err := b.patientRepo.FindByAccessCode(ctx, binding.AccessCode)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Пациент не найден. Запись могла быть удалена.")
		return
	}

	statusText := fmt.Sprintf(
		"Пациент: %s %s\nСтатус: %s\nОперация: %s (%s)",
		patient.FirstName, patient.LastName,
		patient.Status, patient.OperationType, patient.Eye,
	)

	if patient.SurgeryDate != nil {
		statusText += fmt.Sprintf("\nДата операции: %s", patient.SurgeryDate.Format("02.01.2006"))
	}

	b.sendMessage(msg.Chat.ID, statusText)
}

func (b *Bot) SendNotification(chatID int64, text string) {
	if b == nil || b.api == nil {
		return
	}
	b.sendMessage(chatID, text)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Error().Err(err).Int64("chat_id", chatID).Msg("не удалось отправить Telegram сообщение")
	}
}
