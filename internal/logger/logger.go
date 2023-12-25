package logger

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var once sync.Once

func Init() {
	once.Do(func() {
		output := zerolog.ConsoleWriter{Out: os.Stdout}

		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = zerolog.New(output).With().Timestamp().Logger()
	})
}

// GetLogger возвращает глобальный экземпляр логгера
func GetLogger() *zerolog.Logger {
	return &log.Logger
}
