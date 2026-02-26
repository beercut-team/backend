package telegram

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type Bot struct {
	api          *tgbotapi.BotAPI
	patientRepo  repository.PatientRepository
	telegramRepo repository.TelegramRepository
	tokenRepo    repository.TelegramTokenRepository
	userRepo     repository.UserRepository
	baseURL      string
}

func NewBot(token string, baseURL string, patientRepo repository.PatientRepository, telegramRepo repository.TelegramRepository, tokenRepo repository.TelegramTokenRepository, userRepo repository.UserRepository) (*Bot, error) {
	if token == "" {
		return nil, nil
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать Telegram бот: %w", err)
	}

	log.Info().Str("bot", api.Self.UserName).Msg("Telegram бот авторизован")

	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &Bot{
		api:          api,
		patientRepo:  patientRepo,
		telegramRepo: telegramRepo,
		tokenRepo:    tokenRepo,
		userRepo:     userRepo,
		baseURL:      baseURL,
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

			// Recover from panics to keep bot running
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error().Interface("panic", r).Int64("chat_id", update.Message.Chat.ID).Msg("паника в обработчике Telegram сообщения")
					}
				}()
				b.handleMessage(update.Message)
			}()
		}
	}()

	log.Info().Msg("Telegram бот слушает обновления")
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	ctx := context.Background()
	text := strings.TrimSpace(msg.Text)

	// Log all incoming messages for debugging
	log.Info().Int64("chat_id", msg.Chat.ID).Str("text", text).Msg("получено Telegram сообщение")

	switch {
	case strings.HasPrefix(text, "/start"):
		b.handleStart(ctx, msg)
	case strings.HasPrefix(text, "/register"):
		b.handleRegisterDoctor(ctx, msg)
	case text == "/status":
		b.handleStatus(ctx, msg)
	case text == "/mypatients":
		b.handleMyPatients(ctx, msg)
	case text == "/rebind" || text == "/unbind":
		b.handleRebind(ctx, msg)
	case text == "/login":
		b.handleLogin(ctx, msg)
	case text == "/help":
		b.sendMessage(msg.Chat.ID, `Доступные команды:

Для пациентов:
/start <код_доступа> — Привязать к карте пациента
/status — Проверить статус подготовки
/login — Получить ссылку для входа в личный кабинет
/rebind — Отвязать текущего пациента и привязать нового

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

	// Normalize access code: trim whitespace and convert to lowercase
	accessCode := strings.ToLower(strings.TrimSpace(parts[1]))
	log.Info().Str("access_code", accessCode).Int64("chat_id", msg.Chat.ID).Msg("Попытка привязки пациента")

	patient, err := b.patientRepo.FindByAccessCode(ctx, accessCode)
	if err != nil {
		log.Error().Err(err).Str("access_code", accessCode).Msg("Код доступа не найден")
		b.sendMessage(msg.Chat.ID, "Неверный код доступа. Пожалуйста, проверьте и попробуйте снова.")
		return
	}

	// Check if already bound - if so, deactivate old binding
	existing, _ := b.telegramRepo.FindByChatID(ctx, msg.Chat.ID)
	if existing != nil {
		log.Info().Uint("old_patient_id", existing.PatientID).Uint("new_patient_id", patient.ID).Msg("Перепривязка чата")
	}

	binding := &domain.TelegramBinding{
		PatientID:  patient.ID,
		ChatID:     msg.Chat.ID,
		AccessCode: accessCode,
		IsActive:   true,
	}

	// Use UpdateOrCreate to handle both new bindings and rebindings
	if err := b.telegramRepo.UpdateOrCreate(ctx, binding); err != nil {
		log.Error().Err(err).Msg("Не удалось создать/обновить привязку")
		b.sendMessage(msg.Chat.ID, fmt.Sprintf("Не удалось привязать: %v", err))
		return
	}

	statusName := domain.GetStatusDisplayName(patient.Status)
	b.sendMessage(msg.Chat.ID, fmt.Sprintf(
		"✅ Успешно привязано!\nПациент: %s %s\nСтатус: %s\n\nИспользуйте /status для проверки прогресса подготовки.",
		patient.FirstName, patient.LastName, statusName,
	))
}

func (b *Bot) handleStatus(ctx context.Context, msg *tgbotapi.Message) {
	binding, err := b.telegramRepo.FindByChatID(ctx, msg.Chat.ID)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "❌ Вы не привязаны к пациенту.\n\nИспользуйте /start <код_доступа> для привязки.\nКод доступа можно получить у вашего врача.")
		return
	}

	patient, err := b.patientRepo.FindByAccessCode(ctx, binding.AccessCode)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "❌ Пациент не найден. Запись могла быть удалена.\n\nИспользуйте /rebind для привязки к новому пациенту.")
		return
	}

	// Use human-readable status name
	statusName := domain.GetStatusDisplayName(patient.Status)
	operationName := domain.GetOperationTypeDisplayName(patient.OperationType)
	eyeName := domain.GetEyeDisplayName(patient.Eye)

	statusText := fmt.Sprintf(
		"📋 Информация о пациенте\n\nПациент: %s %s\nСтатус: %s\nОперация: %s (%s)",
		patient.FirstName, patient.LastName,
		statusName, operationName, eyeName,
	)

	if patient.SurgeryDate != nil {
		statusText += fmt.Sprintf("\nДата операции: %s", patient.SurgeryDate.Format("02.01.2006"))
	}

	b.sendMessage(msg.Chat.ID, statusText)
}

func (b *Bot) handleRebind(ctx context.Context, msg *tgbotapi.Message) {
	// Check if there's an existing binding
	existing, err := b.telegramRepo.FindByChatID(ctx, msg.Chat.ID)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "У вас нет активной привязки.\n\nИспользуйте /start <код_доступа> для привязки к пациенту.")
		return
	}

	// Deactivate the existing binding
	if err := b.telegramRepo.Delete(ctx, msg.Chat.ID); err != nil {
		log.Error().Err(err).Msg("Не удалось деактивировать привязку")
		b.sendMessage(msg.Chat.ID, fmt.Sprintf("Произошла ошибка: %v", err))
		return
	}

	log.Info().Uint("patient_id", existing.PatientID).Int64("chat_id", msg.Chat.ID).Msg("Привязка деактивирована")

	b.sendMessage(msg.Chat.ID, "✅ Привязка отменена.\n\nТеперь используйте /start <код_доступа> для привязки к новому пациенту.")
}

func (b *Bot) handleLogin(ctx context.Context, msg *tgbotapi.Message) {
	if b.tokenRepo == nil {
		b.sendMessage(msg.Chat.ID, "Функция входа недоступна.")
		return
	}

	// Check if user has an active binding
	binding, err := b.telegramRepo.FindByChatID(ctx, msg.Chat.ID)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "❌ Вы не привязаны к пациенту.\n\nСначала используйте /start <код_доступа> для привязки.")
		return
	}

	// Generate random token (32 characters)
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Error().Err(err).Msg("Не удалось сгенерировать токен")
		b.sendMessage(msg.Chat.ID, fmt.Sprintf("Произошла ошибка при генерации токена: %v", err))
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Create token record with 15-minute expiration
	loginToken := &domain.TelegramLoginToken{
		Token:     token,
		PatientID: binding.PatientID,
		Used:      false,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := b.tokenRepo.Create(ctx, loginToken); err != nil {
		log.Error().Err(err).Msg("Не удалось создать токен входа")
		b.sendMessage(msg.Chat.ID, fmt.Sprintf("Произошла ошибка при создании токена: %v", err))
		return
	}

	// Get patient info
	patient, err := b.patientRepo.FindByID(ctx, binding.PatientID)
	if err != nil {
		log.Error().Err(err).Msg("Не удалось получить информацию о пациенте")
		b.sendMessage(msg.Chat.ID, fmt.Sprintf("Не удалось получить информацию о пациенте: %v", err))
		return
	}

	// Send login link
	loginURL := fmt.Sprintf("%s/patient/portal?token=%s", b.baseURL, token)
	message := fmt.Sprintf(
		"🔐 Вход в личный кабинет\n\nПациент: %s %s\n\n"+
			"Нажмите на ссылку ниже для входа:\n%s\n\n"+
			"⚠️ Ссылка действительна 15 минут и может быть использована только один раз.",
		patient.FirstName, patient.LastName, loginURL,
	)

	b.sendMessage(msg.Chat.ID, message)
	log.Info().Uint("patient_id", binding.PatientID).Str("token", token).Msg("Создан токен входа")
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
		log.Error().Err(err).Msg("Не удалось обновить пользователя")
		b.sendMessage(msg.Chat.ID, fmt.Sprintf("Не удалось привязать аккаунт: %v", err))
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
		statusName := domain.GetStatusDisplayName(p.Status)
		text += fmt.Sprintf("%d. %s %s - %s\n", i+1, p.FirstName, p.LastName, statusName)
	}

	if len(patients) == 10 {
		text += "\n(Показаны первые 10 пациентов)"
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

	// Use human-readable status name
	statusName := domain.GetStatusDisplayName(domain.PatientStatus(newStatus))

	// Status-specific emoji and message
	statusEmoji := map[domain.PatientStatus]string{
		domain.PatientStatusInProgress:      "📝",
		domain.PatientStatusPendingReview:   "👨‍⚕️",
		domain.PatientStatusApproved:        "✅",
		domain.PatientStatusNeedsCorrection: "⚠️",
		domain.PatientStatusScheduled:       "📅",
		domain.PatientStatusCompleted:       "🎉",
		domain.PatientStatusCancelled:       "❌",
	}

	emoji := statusEmoji[domain.PatientStatus(newStatus)]
	if emoji == "" {
		emoji = "🔔"
	}

	operationName := domain.GetOperationTypeDisplayName(patient.OperationType)
	eyeName := domain.GetEyeDisplayName(patient.Eye)

	message := fmt.Sprintf("%s Статус изменён\n\n%s\n\nПациент: %s %s\nОперация: %s (%s)",
		emoji, statusName, patient.FirstName, patient.LastName, operationName, eyeName)

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
		log.Error().Err(err).Uint("patient_id", patientID).Msg("не удалось найти пациента для уведомления хирурга")
		return
	}

	// Найти всех хирургов с привязанным Telegram
	surgeons, err := b.userRepo.FindAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("не удалось найти хирургов")
		return
	}

	districtName := "не указан"
	if patient.District != nil {
		districtName = patient.District.Name
	}

	operationName := domain.GetOperationTypeDisplayName(patient.OperationType)
	eyeName := domain.GetEyeDisplayName(patient.Eye)

	message := fmt.Sprintf("🔍 Требуется проверка\n\nПациент: %s %s\nОперация: %s (%s)\nРайон: %s\n\nИспользуйте веб-интерфейс для проверки документов.",
		patient.FirstName, patient.LastName, operationName, eyeName, districtName)

	sentCount := 0
	for _, surgeon := range surgeons {
		if surgeon.Role == domain.RoleSurgeon && surgeon.TelegramChatID != nil {
			b.sendMessage(*surgeon.TelegramChatID, message)
			sentCount++
		}
	}

	log.Info().Uint("patient_id", patientID).Int("surgeons_notified", sentCount).Msg("уведомления хирургам отправлены")
}

// NotifyPatientNewAccessCode уведомляет пациента о новом коде доступа
func (b *Bot) NotifyPatientNewAccessCode(ctx context.Context, patientID uint, newCode string) {
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

	message := fmt.Sprintf("🔑 Новый код доступа\n\nПациент: %s %s\n\nВаш новый код доступа: %s\n\nИспользуйте его для проверки статуса на сайте: /patient?code=%s",
		patient.FirstName, patient.LastName, newCode, newCode)

	b.sendMessage(binding.ChatID, message)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Error().Err(err).Int64("chat_id", chatID).Msg("не удалось отправить Telegram сообщение")
	}
}
