package database

import (
	"fmt"

	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	log.Info().Msg("подключено к PostgreSQL")

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
		return nil, fmt.Errorf("не удалось выполнить миграцию: %w", err)
	}

	log.Info().Msg("миграция базы данных завершена")
	return db, nil
}
