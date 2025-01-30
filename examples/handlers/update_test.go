package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func ExampleMetricHandler_Update() {
	log := testutils.GetTLogger()
	memStorage := storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(memStorage, log)

	newRouter := chi.NewRouter()
	newRouter.Post("/metrics/update/{metricType}/{metricName}/{value}", mHandlers.Update)

	req := httptest.NewRequest(http.MethodPost, "/metrics/update/gauge/cpu_usage/42", nil)
	response := httptest.NewRecorder()

	newRouter.ServeHTTP(response, req)

	result := response.Result()

	defer result.Body.Close()

	// Print status code
	fmt.Println(result.StatusCode)

	// Output:
	// 200
}
