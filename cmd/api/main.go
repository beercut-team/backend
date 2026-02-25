package main

import (
	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/beercut-team/backend-boilerplate/internal/server"
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

	r := server.NewRouter(cfg, db)

	log.Info().Str("port", cfg.AppPort).Msg("запуск сервера")
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal().Err(err).Msg("сервер остановлен с ошибкой")
	}
}
