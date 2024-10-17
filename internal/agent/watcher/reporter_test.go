package watcher_test

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/stats"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

// Test for SendMetrics function.
func TestMetricService_SendMetrics(t *testing.T) {
	t.Parallel()

	var serverStorage storage.Repository = storage.NewMemStorage()
	r := chi.NewRouter()
	handlers.NewMetricsHandler(serverStorage, r).SetRouter()

	server := httptest.NewServer(r)
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

	collector := watcher.NewMetricCollector(&metrics, &mux, cfg)
	reporter := watcher.NewMetricReporter(&metrics, &mux, cfg)
	collector.UpdateMetrics()

	// Run the SendMetrics function
	reporter.SendMetrics(context.TODO())

	// Compare values on client and on server
	for _, metric := range metrics {
		switch metric.MType {
		case domain.Gauge:
			val, ok := serverStorage.GetGauge(metric.ID)
			assert.True(t, ok)
			original := *(metric.Value)
			assert.InDelta(t, original, val, 00000.1)
		case domain.Counter:
			val, ok := serverStorage.GetCounter(metric.ID)
			assert.True(t, ok)
			original := *(metric.Delta)
			assert.Equal(t, original, val)
		}
	}

	reporter.SendMetrics(context.TODO())
	reporter.SendMetrics(context.TODO())
	counter, ok := serverStorage.GetCounter(domain.PollCount)
	assert.True(t, ok)
	assert.Equal(t, int64(3), counter)
}

func TestMetricReporter_StartReporter(t *testing.T) {
	t.Parallel()

	var serverStorage storage.Repository = storage.NewMemStorage()
	r := chi.NewRouter()
	handlers.NewMetricsHandler(serverStorage, r).SetRouter()

	server := httptest.NewServer(r)
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

	collector := watcher.NewMetricCollector(&metrics, &mux, cfg)
	reporter := watcher.NewMetricReporter(&metrics, &mux, cfg)
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

	counter, ok := serverStorage.GetCounter(domain.PollCount)
	assert.True(t, ok)
	assert.Equal(t, int64(1), counter)

	assert.Equal(t, context.Canceled, ctx.Err())
}
