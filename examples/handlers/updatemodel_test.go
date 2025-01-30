package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func ExampleMetricHandler_UpdateModels() {
	log := testutils.GetTLogger()
	memStorage := storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(memStorage, log)

	jsonPayload := `[
		{"id": "cpu_usage", "type": "gauge", "value": 50},
		{"id": "request_count", "type": "counter", "delta": 10}
	]`

	req := httptest.NewRequest(http.MethodPost, "/metrics/update", strings.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	mHandlers.UpdateModels(response, req)

	result := response.Result()

	defer result.Body.Close()

	// Print status code
	fmt.Println(result.StatusCode)

	// Output:
	// 200
}
