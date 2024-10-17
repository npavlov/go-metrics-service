package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"io"
	"os"
	"sync"
	"time"
)

var (
	once     sync.Once
	instance *zerolog.Logger
)

func Get() *zerolog.Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		log := zerolog.New(output).With().Timestamp().Logger()
		instance = &log
	})
	return instance
}

func SetLogLevel() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}
