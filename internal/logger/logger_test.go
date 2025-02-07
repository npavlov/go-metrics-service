package logger_test

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/logger"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	log := logger.NewLogger()
	require.NotNil(t, log)
	assert.NotNil(t, log.Get())
}

func TestLogger_SetLogLevel(t *testing.T) {
	t.Parallel()

	log := logger.NewLogger()
	log.SetLogLevel(zerolog.DebugLevel)

	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
}

func TestLogger_Output(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.NewLogger()
	newLog := log.Get().Output(&buf)
	newLog.Info().Msg("test message")

	output := buf.String()
	assert.Contains(t, output, "test message")
}
