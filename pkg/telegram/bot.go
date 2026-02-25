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
	userRepo     repository.UserRepository
}

func NewBot(token string, patientRepo repository.PatientRepository, telegramRepo repository.TelegramRepository, userRepo repository.UserRepository) (*Bot, error) {
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
		userRepo:     userRepo,
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
	case strings.HasPrefix(text, "/register"):
		b.handleRegisterDoctor(ctx, msg)
	case text == "/status":
		b.handleStatus(ctx, msg)
	case text == "/mypatients":
		b.handleMyPatients(ctx, msg)
	case text == "/help":
		b.sendMessage(msg.Chat.ID, `Доступные команды:

Для пациентов:
/start <код_доступа> — Привязать к карте пациента
/status — Проверить статус подготовки

Для врачей:
/register <email> — Привязать аккаунт врача
/mypatients — Список моих пациентов
/help — Показать эту справку`)
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

func (b *Bot) handleRegisterDoctor(ctx context.Context, msg *tgbotapi.Message) {
	parts := strings.Fields(msg.Text)
	if len(parts) < 2 {
		b.sendMessage(msg.Chat.ID, "Пожалуйста, укажите ваш email: /register <email>")
		return
	}

	email := parts[1]
	user, err := b.userRepo.FindByEmail(ctx, email)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Пользователь с таким email не найден. Обратитесь к администратору.")
		return
	}

	if user.Role != domain.RoleDistrictDoctor && user.Role != domain.RoleSurgeon && user.Role != domain.RoleAdmin {
		b.sendMessage(msg.Chat.ID, "Регистрация доступна только для врачей и хирургов.")
		return
	}

	chatID := msg.Chat.ID
	user.TelegramChatID = &chatID
	if err := b.userRepo.Update(ctx, user); err != nil {
		b.sendMessage(msg.Chat.ID, "Не удалось привязать аккаунт. Попробуйте позже.")
		return
	}

	roleName := map[domain.Role]string{
		domain.RoleDistrictDoctor: "Районный врач",
		domain.RoleSurgeon:        "Хирург",
		domain.RoleAdmin:          "Администратор",
	}[user.Role]

	b.sendMessage(msg.Chat.ID, fmt.Sprintf(
		"✅ Успешно привязано!\n\nИмя: %s\nРоль: %s\n\nВы будете получать уведомления о новых пациентах и изменениях статусов.",
		user.Name, roleName,
	))
}

func (b *Bot) handleMyPatients(ctx context.Context, msg *tgbotapi.Message) {
	user, err := b.userRepo.FindByChatID(ctx, msg.Chat.ID)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "Аккаунт не привязан. Используйте /register <email>")
		return
	}

	filters := repository.PatientFilters{}
	if user.Role == domain.RoleDistrictDoctor {
		filters.DoctorID = &user.ID
	}

	patients, _, err := b.patientRepo.FindAll(ctx, filters, 0, 10)
	if err != nil || len(patients) == 0 {
		b.sendMessage(msg.Chat.ID, "У вас пока нет пациентов.")
		return
	}

	text := "📋 Ваши пациенты:\n\n"
	for i, p := range patients {
		text += fmt.Sprintf("%d. %s %s - %s\n", i+1, p.FirstName, p.LastName, p.Status)
	}

	b.sendMessage(msg.Chat.ID, text)
}

func (b *Bot) SendNotification(chatID int64, text string) {
	if b == nil || b.api == nil {
		return
	}
	b.sendMessage(chatID, text)
}

// NotifyPatientStatusChange отправляет уведомление пациенту об изменении статуса
func (b *Bot) NotifyPatientStatusChange(ctx context.Context, patientID uint, newStatus string) {
	if b == nil || b.api == nil {
		return
	}

	binding, err := b.telegramRepo.FindByPatientID(ctx, patientID)
	if err != nil || !binding.IsActive {
		return
	}

	patient, err := b.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return
	}

	statusText := map[string]string{
		"PREPARATION":   "📝 Идёт подготовка к операции",
		"REVIEW_NEEDED": "👨‍⚕️ Документы отправлены на проверку хирургу",
		"APPROVED":      "✅ Вы готовы к операции! Ожидайте назначения даты",
		"REJECTED":      "❌ Требуется дополнительная подготовка",
		"SCHEDULED":     "📅 Операция запланирована",
	}[newStatus]

	message := fmt.Sprintf("🔔 Статус изменён\n\n%s\n\nПациент: %s %s\nОперация: %s (%s)",
		statusText, patient.FirstName, patient.LastName, patient.OperationType, patient.Eye)

	if patient.SurgeryDate != nil {
		message += fmt.Sprintf("\n\n📅 Дата операции: %s", patient.SurgeryDate.Format("02.01.2006"))
	}

	b.sendMessage(binding.ChatID, message)
}

// NotifyDoctorNewPatient уведомляет врача о новом пациенте
func (b *Bot) NotifyDoctorNewPatient(ctx context.Context, doctorID uint, patientName string) {
	if b == nil || b.api == nil {
		return
	}

	doctor, err := b.userRepo.FindByID(ctx, doctorID)
	if err != nil || doctor.TelegramChatID == nil {
		return
	}

	message := fmt.Sprintf("👤 Новый пациент\n\n%s добавлен в вашу базу.\nИспользуйте /mypatients для просмотра списка.", patientName)
	b.sendMessage(*doctor.TelegramChatID, message)
}

// NotifySurgeonReviewNeeded уведомляет хирурга о пациенте, готовом к проверке
func (b *Bot) NotifySurgeonReviewNeeded(ctx context.Context, patientID uint) {
	if b == nil || b.api == nil {
		return
	}

	patient, err := b.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return
	}

	// Найти всех хирургов с привязанным Telegram
	surgeons, err := b.userRepo.FindAll(ctx)
	if err != nil {
		return
	}

	message := fmt.Sprintf("🔍 Требуется проверка\n\nПациент: %s %s\nОперация: %s (%s)\nРайон: %s",
		patient.FirstName, patient.LastName, patient.OperationType, patient.Eye, patient.District.Name)

	for _, surgeon := range surgeons {
		if surgeon.Role == domain.RoleSurgeon && surgeon.TelegramChatID != nil {
			b.sendMessage(*surgeon.TelegramChatID, message)
		}
	}
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Error().Err(err).Int64("chat_id", chatID).Msg("не удалось отправить Telegram сообщение")
	}
}
