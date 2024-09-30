package metrics

import (
	"github.com/npavlov/go-metrics-service/internal/server/handlers/update"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

// Test for SendMetrics function
func TestMetricService_SendMetrics(t *testing.T) {

	var serverStorage storage.Repository = storage.NewMemStorage()
	handlers := types.Handlers{
		UpdateHandler:   update.GetUpdateHandler(serverStorage),
		RetrieveHandler: nil,
		RenderHandler:   nil,
	}

	r := router.GetRouter(handlers)

	server := httptest.NewServer(r)
	defer server.Close()

	// Create an instance of MetricService
	var clientStorage storage.Repository = storage.NewMemStorage()
	var collector = NewMetricCollector(clientStorage)
	var reporter = NewMetricReporter(clientStorage, server.URL)
	collector.UpdateMetrics()

	// Run the SendMetrics function
	reporter.SendMetrics()

	serverGauges := serverStorage.GetGauges()
	serverCounters := serverStorage.GetCounters()

	// Compare values on client and on server
	assert.Equal(t, clientStorage.GetGauges(), serverGauges)
	assert.Equal(t, clientStorage.GetCounters(), serverCounters)

	reporter.SendMetrics()
	reporter.SendMetrics()
	counter, ok := serverStorage.GetCounter(types.PollCount)
	assert.True(t, ok)
	assert.Equal(t, int64(3), counter)
}
