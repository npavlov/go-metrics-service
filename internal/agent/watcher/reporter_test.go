package watcher

import (
	"github.com/go-chi/chi/v5"
	"github.com/npavlov/go-metrics-service/internal/agent/stats"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"sync"
	"testing"
)

// Test for SendMetrics function
func TestMetricService_SendMetrics(t *testing.T) {
	var serverStorage storage.Repository = storage.NewMemStorage()
	var r = chi.NewRouter()
	handlers.NewMetricsHandler(serverStorage, r)

	server := httptest.NewServer(r)
	defer server.Close()

	// Create an instance of MetricService
	st := stats.Stats{}
	metrics := st.StatsToMetrics()
	mux := sync.RWMutex{}

	var collector = NewMetricCollector(&metrics, &mux)
	var reporter = NewMetricReporter(&metrics, &mux, server.URL)
	collector.UpdateMetrics()

	// Run the SendMetrics function
	reporter.SendMetrics()

	// Compare values on client and on server
	for _, metric := range metrics {
		switch metric.MType {
		case domain.Gauge:
			val, ok := serverStorage.GetGauge(metric.ID)
			assert.True(t, ok)
			original := *(metric.Value)
			assert.Equal(t, original, val)
		case domain.Counter:
			val, ok := serverStorage.GetCounter(metric.ID)
			assert.True(t, ok)
			original := *(metric.Counter)
			assert.Equal(t, original, val)

		}
	}

	reporter.SendMetrics()
	reporter.SendMetrics()
	counter, ok := serverStorage.GetCounter(domain.PollCount)
	assert.True(t, ok)
	assert.Equal(t, int64(3), counter)
}
