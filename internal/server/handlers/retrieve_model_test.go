package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/stretchr/testify/require"

	"github.com/go-chi/chi/v5"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestUpdateRetrieveModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		update           *model.Metric
		retrieve         *model.Metric
		expectedCode     int
		expectedResponse *model.Metric
	}{
		{
			name: "Successful Gauge Retrieval",
			update: &model.Metric{
				ID:    "Alloc",
				MType: "gauge",
				Value: float64Ptr(100.0),
			},
			retrieve: &model.Metric{
				ID:    "Alloc",
				MType: "gauge",
			},
			expectedCode: http.StatusOK,
			expectedResponse: &model.Metric{
				ID:    "Alloc",
				MType: "gauge",
				Value: float64Ptr(100.0),
			},
		},
		{
			name: "Successful Counter Retrieval",
			update: &model.Metric{
				ID:    "RequestCount",
				MType: "counter",
				Delta: int64Ptr(42),
			},
			retrieve: &model.Metric{
				ID:    "RequestCount",
				MType: "counter",
			},
			expectedCode: http.StatusOK,
			expectedResponse: &model.Metric{
				ID:    "RequestCount",
				MType: "counter",
				Delta: int64Ptr(42),
			},
		},
		{
			name:         "Invalid JSON Payload on Update",
			update:       nil,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Retrieve Non-Existent Metric",
			retrieve: &model.Metric{
				ID:    "NonExistent",
				MType: "gauge",
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "Invalid Metric Type",
			update: &model.Metric{
				ID:    "InvalidTypeMetric",
				MType: "invalidType",
				Value: float64Ptr(1.0),
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Initialize storage and router
			memStorage := storage.NewMemStorage()
			r := chi.NewRouter()
			hd := handlers.NewMetricsHandler(memStorage, r)
			hd.SetRouter()

			// Start the test server
			server := httptest.NewServer(r)
			defer server.Close()

			// Run the update and retrieve tests
			if tt.update != nil {
				testUpdateModel(t, server, tt.update, tt.expectedCode)
			}
			if tt.retrieve != nil {
				testRetrieveModel(t, server, tt.retrieve, tt.expectedCode, tt.expectedResponse)
			}
		})
	}
}

// testUpdateModel handles sending an update request and validating the response.
func testUpdateModel(t *testing.T, server *httptest.Server, request *model.Metric, expectedCode int) {
	t.Helper()

	payload, err := json.Marshal(request)
	require.NoError(t, err)

	res, statusCode, err := sendRequest(t, server, "/update/", payload)
	assert.Equal(t, expectedCode, statusCode)

	if expectedCode == http.StatusOK {
		assert.Equal(t, request, res)
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}
}

// testRetrieveModel handles sending a retrieval request and validating the response.
func testRetrieveModel(t *testing.T, server *httptest.Server, request *model.Metric, expectedCode int, expectedResponse *model.Metric) {
	t.Helper()

	payload, err := json.Marshal(request)
	require.NoError(t, err)

	res, statusCode, err := sendRequest(t, server, "/value/", payload)

	assert.Equal(t, expectedCode, statusCode)

	if expectedCode == http.StatusOK {
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, res)
	} else {
		require.Error(t, err)
	}
}

// sendRequest simplifies sending requests to the test server.
func sendRequest(t *testing.T, server *httptest.Server, route string, payload interface{}) (*model.Metric, int, error) {
	t.Helper()

	url := server.URL + route
	client := resty.New()
	resp, err := client.R().SetHeader("Content-Type", "application/json").SetBody(payload).Post(url)
	if err != nil {
		return nil, resp.StatusCode(), err
	}

	var metric *model.Metric
	if err := json.Unmarshal(resp.Body(), &metric); err != nil {
		return nil, resp.StatusCode(), err
	}

	return metric, resp.StatusCode(), nil
}

// Helper function to create float64 pointer.
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create int64 pointer.
func int64Ptr(i int64) *int64 {
	return &i
}
