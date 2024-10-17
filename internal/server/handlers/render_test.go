package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetRenderHandler(t *testing.T) {
	var memStorage storage.Repository = storage.NewMemStorage()
	var r = chi.NewRouter()
	NewMetricsHandler(memStorage, r)

	// Sample data to return from the mock repository
	gauges := map[domain.MetricName]string{
		"GaugeMetric1": "123.45",
		"GaugeMetric2": "678.90",
	}
	counters := map[domain.MetricName]string{
		"CounterMetric1": "100",
		"CounterMetric2": "200",
	}

	for k, v := range gauges {
		err := memStorage.UpdateMetric(domain.Gauge, k, v)
		assert.Nil(t, err)
	}

	for k, v := range counters {
		err := memStorage.UpdateMetric(domain.Counter, k, v)
		assert.Nil(t, err)
	}

	server := httptest.NewServer(r)
	defer server.Close()

	req := resty.New().R()
	req.Method = http.MethodGet
	req.URL = server.URL

	res, err := req.Send()

	assert.NoError(t, err)
	// Check the status code
	assert.Equal(t, http.StatusOK, res.StatusCode())

	// Verify that the template was rendered correctly by checking if the response body contains the expected data
	body := string(res.Body())

	assert.True(t, strings.Contains(body, "GaugeMetric1"))
	assert.True(t, strings.Contains(body, "123.45"))
	assert.True(t, strings.Contains(body, "GaugeMetric2"))
	assert.True(t, strings.Contains(body, "678.9"))

	assert.True(t, strings.Contains(body, "CounterMetric1"))
	assert.True(t, strings.Contains(body, "100"))
	assert.True(t, strings.Contains(body, "CounterMetric2"))
	assert.True(t, strings.Contains(body, "200"))
}
