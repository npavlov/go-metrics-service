package logger

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	mx sync.Mutex
	lg zerolog.Logger
}

func NewLogger() *Logger {
	//nolint:exhaustruct
	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	return &Logger{
		mx: sync.Mutex{},
		lg: zerolog.New(output).With().Timestamp().Logger(),
	}
}

func (l *Logger) Get() *zerolog.Logger {
	return &l.lg
}

// SetLogLevel sets the global log level for all loggers.
func (l *Logger) SetLogLevel(level zerolog.Level) *Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(level)

	return l
}
