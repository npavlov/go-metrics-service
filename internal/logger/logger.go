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

func NewLogger() *Logger {
	//nolint:exhaustruct
	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	return &Logger{
		mx: sync.RWMutex{},
		lg: zerolog.New(output).With().Timestamp().Logger(),
	}
}

func (l *Logger) Get() *zerolog.Logger {
	l.mx.RLock()
	defer l.mx.RUnlock()
	
	return &l.lg
}

func (l *Logger) SetLogLevel(level zerolog.Level) *Logger {
	l.mx.Lock()
	defer l.mx.Unlock()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(level)

	return l
}
