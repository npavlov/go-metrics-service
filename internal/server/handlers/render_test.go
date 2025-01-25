package handlers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/config"

	"github.com/npavlov/go-metrics-service/internal/server/db"

	"github.com/npavlov/go-metrics-service/internal/server/router"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestGetRenderHandler(t *testing.T) {
	t.Parallel()

	log := testutils.GetTLogger()
	memStorage := storage.NewMemStorage(log)
	cfg := config.NewConfigBuilder(log).Build()
	mHandlers := handlers.NewMetricsHandler(memStorage, log)
	var cRouter router.Router = router.NewCustomRouter(cfg, log)
	cRouter.SetRouter(mHandlers, nil)

	metrics := []db.Metric{
		*db.NewMetric("GaugeMetric1", domain.Gauge, nil, float64Ptr(123.45)),
		*db.NewMetric("GaugeMetric2", domain.Gauge, nil, float64Ptr(678.90)),
		*db.NewMetric("CounterMetric1", domain.Counter, int64Ptr(100), nil),
		*db.NewMetric("CounterMetric2", domain.Counter, int64Ptr(200), nil),
	}

	for _, v := range metrics {
		err := memStorage.Update(context.Background(), &v)
		require.NoError(t, err)
	}

	server := httptest.NewServer(cRouter.GetRouter())
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

func ExampleMetricHandler_Render() {
	log := testutils.GetTLogger()
	memStorage := storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(memStorage, log)

	metrics := []db.Metric{
		*db.NewMetric("GaugeMetric1", domain.Gauge, nil, float64Ptr(123.45)),
		*db.NewMetric("GaugeMetric2", domain.Gauge, nil, float64Ptr(678.90)),
		*db.NewMetric("CounterMetric1", domain.Counter, int64Ptr(100), nil),
		*db.NewMetric("CounterMetric2", domain.Counter, int64Ptr(200), nil),
	}

	req := httptest.NewRequest(http.MethodGet, "/render", nil)
	w := httptest.NewRecorder()

	for _, v := range metrics {
		_ = memStorage.Update(context.Background(), &v)
	}

	mHandlers.Render(w, req)

	resp := w.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Print status code
	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}
