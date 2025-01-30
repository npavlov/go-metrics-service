package handlers_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/npavlov/go-metrics-service/internal/agent/model"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func ExampleMetricHandler_RetrieveModel() {
	log := testutils.GetTLogger()
	memStorage := storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(memStorage, log)
	originMetric := db.NewMetric(domain.LastGC, domain.Gauge, nil, float64Ptr(0.001))
	_ = memStorage.Create(context.Background(), originMetric)

	newRouter := chi.NewRouter()
	newRouter.Get("/value", mHandlers.RetrieveModel)

	requestModel := db.NewMetric(domain.LastGC, domain.Gauge, nil, nil)

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	payload, err := json.Marshal(requestModel)

	req := httptest.NewRequest(http.MethodGet, "/value", bytes.NewReader(payload))
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

	returnMetric := &model.Metric{}
	_ = json.Unmarshal(body, returnMetric)

	// Print status code
	fmt.Println(result.StatusCode)
	// Print response body as string
	fmt.Println(returnMetric.ID, returnMetric.MType, *returnMetric.Value)

	// Output:
	// 200
	// LastGC gauge 0.001
}
