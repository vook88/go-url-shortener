package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func New(level zerolog.Level) zerolog.Logger {
	return log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(level).With().Timestamp().Logger()
}
