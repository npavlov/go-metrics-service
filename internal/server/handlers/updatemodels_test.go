package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/config"

	"github.com/npavlov/go-metrics-service/internal/server/db"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
)

func TestMetricHandler_UpdateModels(t *testing.T) {
	t.Parallel()

	log := logger.NewLogger().Get()
	memStorage := storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(memStorage, log)
	cfg := config.NewConfigBuilder(log).Build()
	var cRouter router.Router = router.NewCustomRouter(cfg, log)
	cRouter.SetRouter(mHandlers, nil)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	tests := []struct {
		name           string
		inputBody      interface{}
		prepareStorage func()
		expectedStatus int
		expectedBody   []db.Metric
	}{
		{
			name: "Valid metrics - Update existing and create new",
			inputBody: []db.Metric{
				*db.NewMetric("existing_counter", domain.Counter, int64Ptr(20), nil),
				*db.NewMetric("new_gauge", domain.Gauge, nil, float64Ptr(123.45)),
			},
			prepareStorage: func() {
				// Prepopulate storage with "existing_counter"
				_ = memStorage.UpdateMany(context.TODO(), &[]db.Metric{
					*db.NewMetric("existing_counter", domain.Counter, int64Ptr(10), nil),
				})
			},
			expectedStatus: http.StatusOK,
			expectedBody: []db.Metric{
				*db.NewMetric("existing_counter", domain.Counter, int64Ptr(30), nil), // Updated Delta
				*db.NewMetric("new_gauge", domain.Gauge, nil, float64Ptr(123.45)),    // New metric
			},
		},
		{
			name: "Invalid metric type in input",
			inputBody: []map[string]interface{}{
				{"id": "invalid_metric", "type": "invalid_type", "delta": 10},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty input metrics",
			inputBody:      []db.Metric{},
			expectedStatus: http.StatusOK,
			expectedBody:   []db.Metric{},
		},
		{
			name:           "Malformed JSON input",
			inputBody:      `{invalid_json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Valid metrics - Batch processing",
			inputBody: []db.Metric{
				*db.NewMetric("new_counter", domain.Counter, int64Ptr(20), nil),
				*db.NewMetric("new_counter", domain.Counter, int64Ptr(10), nil),
				*db.NewMetric("new_counter", domain.Counter, int64Ptr(70), nil),
			},
			expectedStatus: http.StatusOK,
			expectedBody: []db.Metric{
				*db.NewMetric("new_counter", domain.Counter, int64Ptr(100), nil), // Updated Delta
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.prepareStorage != nil {
				tt.prepareStorage()
			}

			// Convert inputBody to JSON for request
			bodyBytes, err := json.Marshal(tt.inputBody)
			require.NoError(t, err)

			// Create request and response recorder
			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Call UpdateModels handler
			mHandlers.UpdateModels(rec, req)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// If we expect a response body, decode and verify it
			if tt.expectedStatus == http.StatusOK && tt.expectedBody != nil {
				var respMetrics []db.Metric
				err := json.NewDecoder(rec.Body).Decode(&respMetrics)
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedBody, respMetrics)
			}
		})
	}
}
