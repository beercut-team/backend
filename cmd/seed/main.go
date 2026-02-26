package main

import (
	"context"

	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/pkg/database"
	"github.com/beercut-team/backend-boilerplate/pkg/logger"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	logger.Init()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("не удалось загрузить конфигурацию")
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("не удалось подключиться к базе данных")
	}

	ctx := context.Background()

	// --- Seed Districts ---
	districts := []domain.District{
		{Name: "Якутск", Region: "Республика Саха (Якутия)", Code: "YAK-01", Timezone: "Asia/Yakutsk"},
		{Name: "Абыйский улус", Region: "Республика Саха (Якутия)", Code: "YAK-02", Timezone: "Asia/Yakutsk"},
		{Name: "Алданский улус", Region: "Республика Саха (Якутия)", Code: "YAK-03", Timezone: "Asia/Yakutsk"},
		{Name: "Аллаиховский улус", Region: "Республика Саха (Якутия)", Code: "YAK-04", Timezone: "Asia/Yakutsk"},
		{Name: "Амгинский улус", Region: "Республика Саха (Якутия)", Code: "YAK-05", Timezone: "Asia/Yakutsk"},
		{Name: "Анабарский улус", Region: "Республика Саха (Якутия)", Code: "YAK-06", Timezone: "Asia/Yakutsk"},
		{Name: "Булунский улус", Region: "Республика Саха (Якутия)", Code: "YAK-07", Timezone: "Asia/Yakutsk"},
		{Name: "Верхневилюйский улус", Region: "Республика Саха (Якутия)", Code: "YAK-08", Timezone: "Asia/Yakutsk"},
		{Name: "Верхнеколымский улус", Region: "Республика Саха (Якутия)", Code: "YAK-09", Timezone: "Asia/Yakutsk"},
		{Name: "Верхоянский улус", Region: "Республика Саха (Якутия)", Code: "YAK-10", Timezone: "Asia/Yakutsk"},
		{Name: "Вилюйский улус", Region: "Республика Саха (Якутия)", Code: "YAK-11", Timezone: "Asia/Yakutsk"},
		{Name: "Горный улус", Region: "Республика Саха (Якутия)", Code: "YAK-12", Timezone: "Asia/Yakutsk"},
		{Name: "Жиганский улус", Region: "Республика Саха (Якутия)", Code: "YAK-13", Timezone: "Asia/Yakutsk"},
		{Name: "Кобяйский улус", Region: "Республика Саха (Якутия)", Code: "YAK-14", Timezone: "Asia/Yakutsk"},
		{Name: "Ленский улус", Region: "Республика Саха (Якутия)", Code: "YAK-15", Timezone: "Asia/Yakutsk"},
		{Name: "Мегино-Кангаласский улус", Region: "Республика Саха (Якутия)", Code: "YAK-16", Timezone: "Asia/Yakutsk"},
		{Name: "Мирнинский улус", Region: "Республика Саха (Якутия)", Code: "YAK-17", Timezone: "Asia/Yakutsk"},
		{Name: "Момский улус", Region: "Республика Саха (Якутия)", Code: "YAK-18", Timezone: "Asia/Yakutsk"},
		{Name: "Намский улус", Region: "Республика Саха (Якутия)", Code: "YAK-19", Timezone: "Asia/Yakutsk"},
		{Name: "Нерюнгринский улус", Region: "Республика Саха (Якутия)", Code: "YAK-20", Timezone: "Asia/Yakutsk"},
		{Name: "Нижнеколымский улус", Region: "Республика Саха (Якутия)", Code: "YAK-21", Timezone: "Asia/Yakutsk"},
		{Name: "Нюрбинский улус", Region: "Республика Саха (Якутия)", Code: "YAK-22", Timezone: "Asia/Yakutsk"},
		{Name: "Оймяконский улус", Region: "Республика Саха (Якутия)", Code: "YAK-23", Timezone: "Asia/Yakutsk"},
		{Name: "Олёкминский улус", Region: "Республика Саха (Якутия)", Code: "YAK-24", Timezone: "Asia/Yakutsk"},
		{Name: "Оленёкский улус", Region: "Республика Саха (Якутия)", Code: "YAK-25", Timezone: "Asia/Yakutsk"},
		{Name: "Среднеколымский улус", Region: "Республика Саха (Якутия)", Code: "YAK-26", Timezone: "Asia/Yakutsk"},
		{Name: "Сунтарский улус", Region: "Республика Саха (Якутия)", Code: "YAK-27", Timezone: "Asia/Yakutsk"},
		{Name: "Таттинский улус", Region: "Республика Саха (Якутия)", Code: "YAK-28", Timezone: "Asia/Yakutsk"},
		{Name: "Томпонский улус", Region: "Республика Саха (Якутия)", Code: "YAK-29", Timezone: "Asia/Yakutsk"},
		{Name: "Усть-Алданский улус", Region: "Республика Саха (Якутия)", Code: "YAK-30", Timezone: "Asia/Yakutsk"},
		{Name: "Усть-Майский улус", Region: "Республика Саха (Якутия)", Code: "YAK-31", Timezone: "Asia/Yakutsk"},
		{Name: "Усть-Янский улус", Region: "Республика Саха (Якутия)", Code: "YAK-32", Timezone: "Asia/Yakutsk"},
		{Name: "Хангаласский улус", Region: "Республика Саха (Якутия)", Code: "YAK-33", Timezone: "Asia/Yakutsk"},
		{Name: "Чурапчинский улус", Region: "Республика Саха (Якутия)", Code: "YAK-34", Timezone: "Asia/Yakutsk"},
		{Name: "Эвено-Бытантайский улус", Region: "Республика Саха (Якутия)", Code: "YAK-35", Timezone: "Asia/Yakutsk"},
	}

	for i := range districts {
		result := db.WithContext(ctx).Where("code = ?", districts[i].Code).FirstOrCreate(&districts[i])
		if result.Error != nil {
			log.Error().Err(result.Error).Str("district", districts[i].Name).Msg("не удалось добавить район")
		}
	}
	log.Info().Int("count", len(districts)).Msg("районы добавлены")

	// --- Seed Users ---
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("123123123"), bcrypt.DefaultCost)
	testHash, _ := bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.DefaultCost)

	users := []domain.User{
		{
			Email: "doctor1@example.com", PasswordHash: string(hash), Name: "Николаев Айсен",
			FirstName: "Айсен", LastName: "Николаев", MiddleName: "Петрович",
			Phone: "+79001234567", Role: domain.RoleDistrictDoctor,
			DistrictID: &districts[0].ID, Specialization: "Офтальмолог", IsActive: true,
		},
		{
			Email: "doctor2@example.com", PasswordHash: string(hash), Name: "Иванова Сардаана",
			FirstName: "Сардаана", LastName: "Иванова", MiddleName: "Алексеевна",
			Phone: "+79001234568", Role: domain.RoleDistrictDoctor,
			DistrictID: &districts[1].ID, Specialization: "Офтальмолог", IsActive: true,
		},
		{
			Email: "surgeon@example.com", PasswordHash: string(hash), Name: "Васильев Ньургун",
			FirstName: "Ньургун", LastName: "Васильев", MiddleName: "Иванович",
			Phone: "+79001234569", Role: domain.RoleSurgeon,
			Specialization: "Хирург-офтальмолог", LicenseNumber: "ЛИЦ-12345", IsActive: true,
		},
		{
			Email: "admin@example.com", PasswordHash: string(hash), Name: "Администратор",
			FirstName: "Админ", LastName: "Системный", MiddleName: "",
			Phone: "+79001234570", Role: domain.RoleAdmin, IsActive: true,
		},
		{
			Email: "admin@gmail.com", PasswordHash: string(adminHash), Name: "Администратор Панели",
			FirstName: "Админ", LastName: "Панель", MiddleName: "",
			Phone: "+79001234571", Role: domain.RoleAdmin, IsActive: true,
		},
		// Test users for automated testing
		{
			Email: "surgeon@test.com", PasswordHash: string(testHash), Name: "Test Surgeon",
			FirstName: "Test", LastName: "Surgeon", MiddleName: "",
			Phone: "+79991234567", Role: domain.RoleSurgeon,
			Specialization: "Хирург-офтальмолог", LicenseNumber: "TEST-001", IsActive: true,
		},
		{
			Email: "doctor@test.com", PasswordHash: string(testHash), Name: "Test Doctor",
			FirstName: "Test", LastName: "Doctor", MiddleName: "",
			Phone: "+79991234568", Role: domain.RoleDistrictDoctor,
			DistrictID: &districts[0].ID, Specialization: "Офтальмолог", IsActive: true,
		},
		{
			Email: "call@test.com", PasswordHash: string(testHash), Name: "Test Call Center",
			FirstName: "Test", LastName: "CallCenter", MiddleName: "",
			Phone: "+79991234569", Role: domain.RoleCallCenter, IsActive: true,
		},
	}

	for i := range users {
		result := db.WithContext(ctx).Where("email = ?", users[i].Email).FirstOrCreate(&users[i])
		if result.Error != nil {
			log.Error().Err(result.Error).Str("email", users[i].Email).Msg("не удалось добавить пользователя")
		}
	}
	log.Info().Int("count", len(users)).Msg("пользователи добавлены")

	// --- Seed Patients ---
	patients := []domain.Patient{
		{
			AccessCode: "a1b2c3d4", FirstName: "Туяра", LastName: "Алексеева", MiddleName: "Петровна",
			Phone: "+79009876543", Diagnosis: "Катаракта правого глаза, начальная стадия",
			OperationType: domain.OperationPhacoemulsification, Eye: "OD",
			Status: domain.PatientStatusInProgress, DoctorID: users[0].ID, DistrictID: districts[0].ID,
		},
		{
			AccessCode: "e5f6g7h8", FirstName: "Айаал", LastName: "Степанов", MiddleName: "Николаевич",
			Phone: "+79009876544", Diagnosis: "Открытоугольная глаукома II стадии",
			OperationType: domain.OperationAntiglaucoma, Eye: "OS",
			Status: domain.PatientStatusPendingReview, DoctorID: users[0].ID, DistrictID: districts[0].ID,
			SurgeonID: &users[2].ID,
		},
		{
			AccessCode: "i9j0k1l2", FirstName: "Айыына", LastName: "Павлова", MiddleName: "Ивановна",
			Phone: "+79009876545", Diagnosis: "Регматогенная отслойка сетчатки",
			OperationType: domain.OperationVitrectomy, Eye: "OD",
			Status: domain.PatientStatusApproved, DoctorID: users[1].ID, DistrictID: districts[1].ID,
			SurgeonID: &users[2].ID,
		},
	}

	for i := range patients {
		result := db.WithContext(ctx).Where("access_code = ?", patients[i].AccessCode).FirstOrCreate(&patients[i])
		if result.Error != nil {
			log.Error().Err(result.Error).Str("patient", patients[i].LastName).Msg("не удалось добавить пациента")
		}
	}
	log.Info().Int("count", len(patients)).Msg("пациенты добавлены")

	// --- Seed Checklist Items for first patient ---
	var existingCount int64
	db.Model(&domain.ChecklistItem{}).Where("patient_id = ?", patients[0].ID).Count(&existingCount)

	if existingCount == 0 {
		templates := domain.GetChecklistTemplates(patients[0].OperationType)
		var items []domain.ChecklistItem
		for _, t := range templates {
			items = append(items, domain.ChecklistItem{
				PatientID:   patients[0].ID,
				Name:        t.Name,
				Description: t.Description,
				Category:    t.Category,
				IsRequired:  t.IsRequired,
				Status:      domain.ChecklistStatusPending,
			})
		}
		if len(items) > 0 {
			db.WithContext(ctx).Create(&items)
			log.Info().Int("count", len(items)).Msg("чек-лист для пациента 1 добавлен")
		}
	}

	log.Info().Msg("заполнение тестовыми данными завершено успешно")
}
