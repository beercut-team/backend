package main

import (
	"context"
	"fmt"

	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/pkg/database"
	"github.com/beercut-team/backend-boilerplate/pkg/logger"
	"github.com/rs/zerolog/log"
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

	log.Info().Msg("🗑️  Удаление всех таблиц...")

	// Drop all tables in reverse order of dependencies
	tables := []interface{}{
		&domain.SyncQueue{},
		&domain.TelegramBinding{},
		&domain.Notification{},
		&domain.Comment{},
		&domain.Surgery{},
		&domain.IOLCalculation{},
		&domain.Media{},
		&domain.ChecklistItem{},
		&domain.ChecklistTemplate{},
		&domain.PatientStatusHistory{},
		&domain.Patient{},
		&domain.AuditLog{},
		&domain.District{},
		&domain.User{},
	}

	for _, table := range tables {
		if err := db.WithContext(ctx).Migrator().DropTable(table); err != nil {
			log.Warn().Err(err).Msgf("не удалось удалить таблицу %T", table)
		}
	}

	log.Info().Msg("📦 Пересоздание схемы...")

	// Recreate schema
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.District{},
		&domain.AuditLog{},
		&domain.Patient{},
		&domain.PatientStatusHistory{},
		&domain.ChecklistTemplate{},
		&domain.ChecklistItem{},
		&domain.Media{},
		&domain.IOLCalculation{},
		&domain.Surgery{},
		&domain.Comment{},
		&domain.Notification{},
		&domain.TelegramBinding{},
		&domain.SyncQueue{},
	); err != nil {
		log.Fatal().Err(err).Msg("не удалось выполнить миграцию")
	}

	// Remove default constraint from is_required column
	if err := db.Exec("ALTER TABLE checklist_items ALTER COLUMN is_required DROP DEFAULT").Error; err != nil {
		log.Warn().Err(err).Msg("не удалось удалить DEFAULT для is_required (возможно, его уже нет)")
	}

	log.Info().Msg("✅ База данных очищена и готова к заполнению")
	fmt.Println("\nТеперь запустите: go run ./cmd/seed")
}
