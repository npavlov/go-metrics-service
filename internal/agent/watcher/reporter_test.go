package watcher_test

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/stats"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

// Test for SendMetrics function.
func TestMetricService_SendMetrics(t *testing.T) {
	t.Parallel()

	log := testutils.GetTLogger()
	var serverStorage storage.Repository = storage.NewMemStorage(log)
	router := chi.NewRouter()
	handlers.NewMetricsHandler(serverStorage, router, log).SetRouter()

	server := httptest.NewServer(router)
	defer server.Close()

	// Create an instance of MetricService
	st := stats.Stats{}
	metrics := st.StatsToMetrics()
	mux := sync.RWMutex{}
	cfg := &config.Config{
		Address:        server.URL,
		PollInterval:   1,
		ReportInterval: 1,
	}

	collector := watcher.NewMetricCollector(&metrics, &mux, cfg, log)
	reporter := watcher.NewMetricReporter(&metrics, &mux, cfg, log)
	collector.UpdateMetrics()

	// Run the SendMetrics function
	reporter.SendMetrics(context.TODO())

	// Compare values on client and on server
	for _, metric := range metrics {
		switch metric.MType {
		case domain.Gauge:
			m, ok := serverStorage.Get(metric.ID)
			assert.True(t, ok)
			original := *(metric.Value)
			assert.InDelta(t, original, *m.Value, 00000.1)
		case domain.Counter:
			m, ok := serverStorage.Get(metric.ID)
			assert.True(t, ok)
			original := *(metric.Delta)
			assert.Equal(t, original, *m.Delta)
		}
	}

	reporter.SendMetrics(context.TODO())
	reporter.SendMetrics(context.TODO())
	m, ok := serverStorage.Get(domain.PollCount)
	assert.True(t, ok)
	assert.Equal(t, int64(3), *m.Delta)
}

func TestMetricReporter_StartReporter(t *testing.T) {
	t.Parallel()

	log := testutils.GetTLogger()
	var serverStorage storage.Repository = storage.NewMemStorage(log)
	router := chi.NewRouter()
	handlers.NewMetricsHandler(serverStorage, router, log).SetRouter()

	server := httptest.NewServer(router)
	defer server.Close()

	// Create an instance of MetricService
	st := stats.Stats{}
	metrics := st.StatsToMetrics()
	mux := sync.RWMutex{}
	cfg := &config.Config{
		Address:        server.URL,
		PollInterval:   1,
		ReportInterval: 2,
	}

	collector := watcher.NewMetricCollector(&metrics, &mux, cfg, log)
	reporter := watcher.NewMetricReporter(&metrics, &mux, cfg, log)
	collector.UpdateMetrics()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go reporter.StartReporter(ctx, &wg)
	// Wait for a short duration to allow the reporter to run
	time.Sleep(2 * time.Second)
	cancel() // Stop the reporter

	wg.Wait() // Wait for the goroutine to finish

	m, ok := serverStorage.Get(domain.PollCount)
	assert.True(t, ok)
	assert.Equal(t, int64(1), *m.Delta)

	assert.Equal(t, context.Canceled, ctx.Err())
}
