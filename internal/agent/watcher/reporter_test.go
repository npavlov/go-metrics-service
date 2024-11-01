package watcher_test

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/server/router"

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
	serverStorage := storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(serverStorage, log)
	var cRouter router.Router = router.NewCustomRouter(log)
	cRouter.SetRouter(mHandlers, nil)

	server := httptest.NewServer(cRouter.GetRouter())
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

	newConfig := config.NewConfigBuilder(log).FromObj(cfg).Build()
	collector := watcher.NewMetricCollector(&metrics, &mux, newConfig, log)
	reporter := watcher.NewMetricReporter(&metrics, &mux, newConfig, log)
	collector.UpdateMetrics()

	// Run the SendMetrics function
	reporter.SendMetrics(context.TODO())

	// Compare values on client and on server
	for _, metric := range metrics {
		switch metric.MType {
		case domain.Gauge:
			m, ok := serverStorage.Get(context.Background(), metric.ID)
			assert.True(t, ok)
			original := *(metric.Value)
			assert.InDelta(t, original, *m.Value, 00000.1)
		case domain.Counter:
			m, ok := serverStorage.Get(context.Background(), metric.ID)
			assert.True(t, ok)
			original := *(metric.Delta)
			assert.Equal(t, original, *m.Delta)
		}
	}

	reporter.SendMetrics(context.TODO())
	reporter.SendMetrics(context.TODO())
	m, ok := serverStorage.Get(context.Background(), domain.PollCount)
	assert.True(t, ok)
	assert.Equal(t, int64(3), *m.Delta)
}

func TestMetricReporter_StartReporter(t *testing.T) {
	t.Parallel()

	log := testutils.GetTLogger()
	serverStorage := storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(serverStorage, log)
	var cRouter router.Router = router.NewCustomRouter(log)
	cRouter.SetRouter(mHandlers, nil)

	server := httptest.NewServer(cRouter.GetRouter())
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
	newConfig := config.NewConfigBuilder(log).FromObj(cfg).Build()

	collector := watcher.NewMetricCollector(&metrics, &mux, newConfig, log)
	reporter := watcher.NewMetricReporter(&metrics, &mux, newConfig, log)
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

	m, ok := serverStorage.Get(context.Background(), domain.PollCount)
	assert.True(t, ok)
	assert.Equal(t, int64(1), *m.Delta)

	assert.Equal(t, context.Canceled, ctx.Err())
}

func TestMetricReporter_SendMetricsBatch(t *testing.T) {
	t.Parallel()

	// Setup Logger and Test Server
	log := testutils.GetTLogger()
	serverStorage := storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(serverStorage, log)
	var cRouter router.Router = router.NewCustomRouter(log)
	cRouter.SetRouter(mHandlers, nil)

	server := httptest.NewServer(cRouter.GetRouter())
	defer server.Close()

	// Create mock metrics and configurations
	st := stats.Stats{}
	metrics := st.StatsToMetrics()
	mux := sync.RWMutex{}
	cfg := &config.Config{
		Address:        server.URL,
		PollInterval:   1,
		ReportInterval: 2,
		UseBatch:       true,
	}
	newConfig := config.NewConfigBuilder(log).FromObj(cfg).Build()

	// Initialize the MetricReporter
	reporter := watcher.NewMetricReporter(&metrics, &mux, newConfig, log)
	collector := watcher.NewMetricCollector(&metrics, &mux, newConfig, log)
	collector.UpdateMetrics()
	// Run the SendMetricsBatch function
	reporter.SendMetricsBatch(context.TODO())

	// Verify that all metrics were sent in a batch and saved on the server
	for _, metric := range metrics {
		switch metric.MType {
		case domain.Gauge:
			m, ok := serverStorage.Get(context.Background(), metric.ID)
			assert.True(t, ok, "Gauge metric should be present in server storage")
			assert.InDelta(t, *metric.Value, *m.Value, 0.0001, "Gauge metric value should match")
		case domain.Counter:
			m, ok := serverStorage.Get(context.Background(), metric.ID)
			assert.True(t, ok, "Counter metric should be present in server storage")
			assert.Equal(t, *metric.Delta, *m.Delta, "Counter metric delta should match")
		}
	}

	// Send batch multiple times and validate PollCount
	reporter.SendMetricsBatch(context.TODO())
	reporter.SendMetricsBatch(context.TODO())

	// Check that PollCount counter has incremented
	m, ok := serverStorage.Get(context.Background(), domain.PollCount)
	assert.True(t, ok, "PollCount metric should be present in server storage")
	assert.Equal(t, int64(3), *m.Delta, "PollCount metric delta should be 3 after three batches")
}
