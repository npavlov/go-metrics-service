package render

import (
	"github.com/go-resty/resty/v2"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetRenderHandler(t *testing.T) {
	var memStorage storage.Repository = storage.NewMemStorage()

	// Sample data to return from the mock repository
	gauges := map[types.MetricName]float64{
		"GaugeMetric1": 123.45,
		"GaugeMetric2": 678.90,
	}
	counters := map[types.MetricName]int64{
		"CounterMetric1": 100,
		"CounterMetric2": 200,
	}

	for k, v := range gauges {
		memStorage.UpdateGauge(k, v)
	}

	for k, v := range counters {
		memStorage.UpdateCounter(k, v)
	}

	handlers := types.Handlers{
		UpdateHandler:   nil,
		RetrieveHandler: nil,
		RenderHandler:   GetRenderHandler(memStorage),
	}

	r := router.GetRouter(handlers)

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