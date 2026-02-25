package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"},
	).With().Timestamp().Caller().Logger()
}
