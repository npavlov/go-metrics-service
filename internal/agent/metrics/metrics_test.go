package metrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/npavlov/go-metrics-service/internal/agent/metrictypes"
	"github.com/npavlov/go-metrics-service/internal/server/handler"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestMetricService_UpdateMetrics(t *testing.T) {
	st := storage.NewMemStorage()
	metricService := NewMetricService(st, "")

	// Call the method to test
	metricService.UpdateMetrics()

	counters := metricService.Storage.GetCounters()
	gauges := metricService.Storage.GetGauges()

	assert.Equal(t, 1, len(counters))
	assert.Equal(t, 28, len(gauges))

	metricService.UpdateMetrics()

	value, ok := metricService.Storage.GetCounter(metrictypes.PollCount)

	assert.True(t, ok)
	assert.Equal(t, int64(2), value)
}

// Test for SendMetrics function
func TestMetricService_SendMetrics(t *testing.T) {
	var serverStorage storage.Repository = storage.NewMemStorage()
	r := chi.NewRouter()

	updateHandler := handler.GetUpdateHandler(serverStorage)
	router.SetRoutes(r, updateHandler)

	server := httptest.NewServer(r)
	defer server.Close()

	// Create an instance of MetricService
	var clientStorage storage.Repository = storage.NewMemStorage()
	service := NewMetricService(clientStorage, server.URL)
	service.UpdateMetrics()

	// Run the SendMetrics function
	service.SendMetrics()

	serverGauges := serverStorage.GetGauges()
	serverCounters := serverStorage.GetCounters()

	// Compare values on client and on server
	assert.Equal(t, clientStorage.GetGauges(), serverGauges)
	assert.Equal(t, clientStorage.GetCounters(), serverCounters)
}
