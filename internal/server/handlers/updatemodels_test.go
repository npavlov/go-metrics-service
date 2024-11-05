package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	var cRouter router.Router = router.NewCustomRouter(log)
	cRouter.SetRouter(mHandlers, nil)

	tests := []struct {
		name           string
		inputBody      interface{}
		prepareStorage func()
		expectedStatus int
		expectedBody   []db.MtrMetric
	}{
		{
			name: "Valid metrics - Update existing and create new",
			inputBody: []db.MtrMetric{
				{ID: "existing_counter", MType: domain.Counter, Delta: int64Ptr(20)},
				{ID: "new_gauge", MType: domain.Gauge, Value: float64Ptr(123.45)},
			},
			prepareStorage: func() {
				// Prepopulate storage with "existing_counter"
				_ = memStorage.UpdateMany(context.TODO(), &[]db.MtrMetric{
					{ID: "existing_counter", MType: domain.Counter, Delta: int64Ptr(10)},
				})
			},
			expectedStatus: http.StatusOK,
			expectedBody: []db.MtrMetric{
				{ID: "existing_counter", MType: domain.Counter, Delta: int64Ptr(30)}, // Updated Delta
				{ID: "new_gauge", MType: domain.Gauge, Value: float64Ptr(123.45)},    // New metric
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
			inputBody:      []db.MtrMetric{},
			expectedStatus: http.StatusOK,
			expectedBody:   []db.MtrMetric{},
		},
		{
			name:           "Malformed JSON input",
			inputBody:      `{invalid_json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Valid metrics - Batch processing",
			inputBody: []db.MtrMetric{
				{ID: "new_counter", MType: domain.Counter, Delta: int64Ptr(20)},
				{ID: "new_counter", MType: domain.Counter, Delta: int64Ptr(10)},
				{ID: "new_counter", MType: domain.Counter, Delta: int64Ptr(70)},
			},
			expectedStatus: http.StatusOK,
			expectedBody: []db.MtrMetric{
				{ID: "new_counter", MType: domain.Counter, Delta: int64Ptr(100)}, // Updated Delta
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
				var respMetrics []db.MtrMetric
				err := json.NewDecoder(rec.Body).Decode(&respMetrics)
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedBody, respMetrics)
			}
		})
	}
}
