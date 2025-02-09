package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/dbmanager"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestSetupLogger(t *testing.T) {
	t.Parallel()

	log := setupLogger()
	assert.NotNil(t, log)
	assert.Equal(t, zerolog.DebugLevel, log.GetLevel())
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	log := setupLogger()
	cfg := loadConfig(&log)

	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Address)
}

func TestSetupDBManager(t *testing.T) {
	t.Parallel()

	log := setupLogger()
	ctx := context.Background()

	cfg := &config.Config{
		Database: "mock_db_connection_string",
	}

	dbManager := setupDBManager(ctx, cfg, &log)
	assert.NotNil(t, dbManager)
}

func TestSetupStorage(t *testing.T) {
	t.Parallel()

	log := setupLogger()
	ctx := context.Background()

	cfg := &config.Config{}
	dbManager := &dbmanager.DBManager{IsConnected: false}

	storage := setupStorage(ctx, cfg, dbManager, &log)
	assert.NotNil(t, storage)
}

func TestStartServer(t *testing.T) {
	t.Parallel()

	log := setupLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &config.Config{Address: "localhost:8080"}
	dbManager := &dbmanager.DBManager{IsConnected: false}
	storage := setupStorage(ctx, cfg, dbManager, &log)

	go func() {
		startServer(ctx, cfg, storage, dbManager, &log)
	}()

	testutils.SendServerRequest(t, "http://"+cfg.Address, "/update/gauge/MSpanInuse/23360.000000", http.StatusOK)

	// Give it some time to start
	time.Sleep(1 * time.Second)

	metric, ok := storage.Get(ctx, domain.MSpanInuse)

	assert.True(t, ok)
	assert.Equal(t, metric.Value, float64Ptr(23360))

	// Trigger shutdown
	cancel()
}

func float64Ptr(f float64) *float64 {
	return &f
}
