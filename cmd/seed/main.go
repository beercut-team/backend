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
		log.Fatal().Err(err).Msg("failed to load config")
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	ctx := context.Background()

	// --- Seed Districts ---
	districts := []domain.District{
		{Name: "Белгородский район", Region: "Белгородская область", Code: "BEL-01", Timezone: "Europe/Moscow"},
		{Name: "Старооскольский район", Region: "Белгородская область", Code: "BEL-02", Timezone: "Europe/Moscow"},
		{Name: "Губкинский район", Region: "Белгородская область", Code: "BEL-03", Timezone: "Europe/Moscow"},
		{Name: "Шебекинский район", Region: "Белгородская область", Code: "BEL-04", Timezone: "Europe/Moscow"},
		{Name: "Валуйский район", Region: "Белгородская область", Code: "BEL-05", Timezone: "Europe/Moscow"},
		{Name: "Алексеевский район", Region: "Белгородская область", Code: "BEL-06", Timezone: "Europe/Moscow"},
	}

	for i := range districts {
		result := db.WithContext(ctx).Where("code = ?", districts[i].Code).FirstOrCreate(&districts[i])
		if result.Error != nil {
			log.Error().Err(result.Error).Str("district", districts[i].Name).Msg("failed to seed district")
		}
	}
	log.Info().Int("count", len(districts)).Msg("seeded districts")

	// --- Seed Users ---
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	users := []domain.User{
		{
			Email: "doctor1@example.com", PasswordHash: string(hash), Name: "Иванов Иван",
			FirstName: "Иван", LastName: "Иванов", MiddleName: "Петрович",
			Phone: "+79001234567", Role: domain.RoleDistrictDoctor,
			DistrictID: &districts[0].ID, Specialization: "Офтальмолог", IsActive: true,
		},
		{
			Email: "doctor2@example.com", PasswordHash: string(hash), Name: "Петрова Мария",
			FirstName: "Мария", LastName: "Петрова", MiddleName: "Сергеевна",
			Phone: "+79001234568", Role: domain.RoleDistrictDoctor,
			DistrictID: &districts[1].ID, Specialization: "Офтальмолог", IsActive: true,
		},
		{
			Email: "surgeon@example.com", PasswordHash: string(hash), Name: "Сидоров Алексей",
			FirstName: "Алексей", LastName: "Сидоров", MiddleName: "Владимирович",
			Phone: "+79001234569", Role: domain.RoleSurgeon,
			Specialization: "Хирург-офтальмолог", LicenseNumber: "ЛИЦ-12345", IsActive: true,
		},
		{
			Email: "admin@example.com", PasswordHash: string(hash), Name: "Администратор",
			FirstName: "Админ", LastName: "Системный", MiddleName: "",
			Phone: "+79001234570", Role: domain.RoleAdmin, IsActive: true,
		},
	}

	for i := range users {
		result := db.WithContext(ctx).Where("email = ?", users[i].Email).FirstOrCreate(&users[i])
		if result.Error != nil {
			log.Error().Err(result.Error).Str("email", users[i].Email).Msg("failed to seed user")
		}
	}
	log.Info().Int("count", len(users)).Msg("seeded users")

	// --- Seed Patients ---
	patients := []domain.Patient{
		{
			AccessCode: "a1b2c3d4", FirstName: "Ольга", LastName: "Кузнецова", MiddleName: "Ивановна",
			Phone: "+79009876543", Diagnosis: "Катаракта правого глаза, начальная стадия",
			OperationType: domain.OperationPhacoemulsification, Eye: "OD",
			Status: domain.PatientStatusPreparation, DoctorID: users[0].ID, DistrictID: districts[0].ID,
		},
		{
			AccessCode: "e5f6g7h8", FirstName: "Сергей", LastName: "Морозов", MiddleName: "Алексеевич",
			Phone: "+79009876544", Diagnosis: "Открытоугольная глаукома II стадии",
			OperationType: domain.OperationAntiglaucoma, Eye: "OS",
			Status: domain.PatientStatusReviewNeeded, DoctorID: users[0].ID, DistrictID: districts[0].ID,
			SurgeonID: &users[2].ID,
		},
		{
			AccessCode: "i9j0k1l2", FirstName: "Анна", LastName: "Белова", MiddleName: "Михайловна",
			Phone: "+79009876545", Diagnosis: "Регматогенная отслойка сетчатки",
			OperationType: domain.OperationVitrectomy, Eye: "OD",
			Status: domain.PatientStatusApproved, DoctorID: users[1].ID, DistrictID: districts[1].ID,
			SurgeonID: &users[2].ID,
		},
	}

	for i := range patients {
		result := db.WithContext(ctx).Where("access_code = ?", patients[i].AccessCode).FirstOrCreate(&patients[i])
		if result.Error != nil {
			log.Error().Err(result.Error).Str("patient", patients[i].LastName).Msg("failed to seed patient")
		}
	}
	log.Info().Int("count", len(patients)).Msg("seeded patients")

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
			log.Info().Int("count", len(items)).Msg("seeded checklist items for patient 1")
		}
	}

	log.Info().Msg("seeding completed successfully")
}
