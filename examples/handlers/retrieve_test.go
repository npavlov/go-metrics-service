package handlers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func ExampleMetricHandler_Retrieve() {
	log := testutils.GetTLogger()
	memStorage := storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(memStorage, log)
	_ = memStorage.Create(context.Background(), db.NewMetric(domain.LastGC, domain.Gauge, nil, float64Ptr(0.001)))

	newRouter := chi.NewRouter()
	newRouter.Get("/metrics/{metricName}", mHandlers.Retrieve)

	req := httptest.NewRequest(http.MethodGet, "/metrics/LastGC", nil)
	resp := httptest.NewRecorder()

	newRouter.ServeHTTP(resp, req)

	result := resp.Result()

	defer result.Body.Close()

	// Read full response body
	body, err := io.ReadAll(result.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Print status code
	fmt.Println(result.StatusCode)
	// Print response body as string
	fmt.Println(string(body))

	// Output:
	// 200
	// 0.001
}

// Helper function to create float64 pointer.
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create int64 pointer.
func int64Ptr(i int64) *int64 {
	return &i
}
