package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/server/storage"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/dbmanager"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	log := testutils.GetTLogger()
	cfg := loadConfig(log)

	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Address)
}

func TestStartServer(t *testing.T) {
	t.Parallel()

	log := testutils.GetTLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &config.Config{Address: "localhost:8080"}
	dbManager := &dbmanager.DBManager{IsConnected: false}
	memStorage := storage.NewMemStorage(log).WithBackup(ctx, cfg)

	go func() {
		startServer(ctx, cfg, memStorage, dbManager, log)
	}()

	testutils.SendServerRequest(t, "http://"+cfg.Address, "/update/gauge/MSpanInuse/23360.000000", http.StatusOK)

	// Give it some time to start
	time.Sleep(1 * time.Second)

	metric, ok := memStorage.Get(ctx, domain.MSpanInuse)

	assert.True(t, ok)
	assert.Equal(t, metric.Value, float64Ptr(23360))

	// Trigger shutdown
	cancel()
}

func float64Ptr(f float64) *float64 {
	return &f
}
