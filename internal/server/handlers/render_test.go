package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/router"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestGetRenderHandler(t *testing.T) {
	t.Parallel()

	log := testutils.GetTLogger()
	var memStorage storage.Repository = storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(memStorage, log)
	var cRouter router.Router = router.NewCustomRouter(log)
	cRouter.SetRouter(mHandlers)

	metrics := []model.Metric{
		{
			ID:    "GaugeMetric1",
			MType: domain.Gauge,
			Value: float64Ptr(123.45),
		},
		{
			ID:    "GaugeMetric2",
			MType: domain.Gauge,
			Value: float64Ptr(678.90),
		},
		{
			ID:    "CounterMetric1",
			MType: domain.Counter,
			Delta: int64Ptr(100),
		},
		{
			ID:    "CounterMetric2",
			MType: domain.Counter,
			Delta: int64Ptr(200),
		},
	}

	for _, v := range metrics {
		err := memStorage.Update(&v)
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
