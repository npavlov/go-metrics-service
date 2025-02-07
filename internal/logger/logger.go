package logger

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	mx sync.RWMutex
	lg zerolog.Logger
}

func NewLogger(level zerolog.Level) *Logger {
	//nolint:exhaustruct
	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	return &Logger{
		mx: sync.RWMutex{},
		lg: zerolog.New(output).Level(level).With().Timestamp().Logger(),
	}
}

func (l *Logger) Get() zerolog.Logger {
	l.mx.RLock()
	defer l.mx.RUnlock()

	return l.lg
}
