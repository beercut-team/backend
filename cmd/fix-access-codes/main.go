package main

import (
	"context"

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

	// Найти всех пациентов без кода доступа
	var patients []domain.Patient
	if err := db.WithContext(ctx).Where("access_code = '' OR access_code IS NULL").Find(&patients).Error; err != nil {
		log.Fatal().Err(err).Msg("не удалось получить пациентов")
	}

	if len(patients) == 0 {
		log.Info().Msg("все пациенты уже имеют коды доступа")
		return
	}

	log.Info().Int("count", len(patients)).Msg("найдено пациентов без кодов")

	// Сгенерировать коды для каждого
	for i := range patients {
		code := domain.GenerateAccessCode()

		// Проверить уникальность
		var exists int64
		for {
			db.Model(&domain.Patient{}).Where("access_code = ?", code).Count(&exists)
			if exists == 0 {
				break
			}
			code = domain.GenerateAccessCode()
		}

		patients[i].AccessCode = code
		if err := db.WithContext(ctx).Model(&patients[i]).Update("access_code", code).Error; err != nil {
			log.Error().Err(err).Uint("patient_id", patients[i].ID).Msg("не удалось обновить код")
		} else {
			log.Info().Uint("id", patients[i].ID).Str("code", code).Str("name", patients[i].FirstName+" "+patients[i].LastName).Msg("код сгенерирован")
		}
	}

	log.Info().Msg("генерация кодов завершена")
}
