package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRenderHandler(t *testing.T) {
	t.Parallel()

	var memStorage storage.Repository = storage.NewMemStorage()
	r := chi.NewRouter()
	handlers.NewMetricsHandler(memStorage, r).SetRouter()

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
		require.NoError(t, err)
	}

	for k, v := range counters {
		err := memStorage.UpdateMetric(domain.Counter, k, v)
		require.NoError(t, err)
	}

	server := httptest.NewServer(r)
	defer server.Close()

	req := resty.New().R()
	req.Method = http.MethodGet
	req.URL = server.URL

	res, err := req.Send()

	require.NoError(t, err)
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
