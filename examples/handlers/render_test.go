package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

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
	resp := httptest.NewRecorder()

	for _, v := range metrics {
		_ = memStorage.Update(context.Background(), &v)
	}

	mHandlers.Render(resp, req)

	result := resp.Result()

	defer result.Body.Close()

	// Print status code
	fmt.Println(result.StatusCode)

	// Output:
	// 200
}
