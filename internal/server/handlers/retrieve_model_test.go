package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/db"

	"github.com/npavlov/go-metrics-service/internal/server/router"

	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestUpdateRetrieveModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		update           *db.MtrMetric
		retrieve         *db.MtrMetric
		expectedCode     int
		expectedResponse *db.MtrMetric
	}{
		{
			name: "Successful Gauge Retrieval",
			update: &db.MtrMetric{
				ID:    "Alloc",
				MType: "gauge",
				Value: float64Ptr(100.0),
			},
			retrieve: &db.MtrMetric{
				ID:    "Alloc",
				MType: "gauge",
			},
			expectedCode: http.StatusOK,
			expectedResponse: &db.MtrMetric{
				ID:    "Alloc",
				MType: "gauge",
				Value: float64Ptr(100.0),
			},
		},
		{
			name: "Successful Counter Retrieval",
			update: &db.MtrMetric{
				ID:    "RequestCount",
				MType: "counter",
				Delta: int64Ptr(42),
			},
			retrieve: &db.MtrMetric{
				ID:    "RequestCount",
				MType: "counter",
			},
			expectedCode: http.StatusOK,
			expectedResponse: &db.MtrMetric{
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
			retrieve: &db.MtrMetric{
				ID:    "NonExistent",
				MType: "gauge",
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Initialize storage and router
			log := testutils.GetTLogger()
			memStorage := storage.NewMemStorage(log)
			mHandlers := handlers.NewMetricsHandler(memStorage, log)
			var cRouter router.Router = router.NewCustomRouter(log)
			cRouter.SetRouter(mHandlers, nil)

			// Start the test server
			server := httptest.NewServer(cRouter.GetRouter())
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
func testUpdateModel(t *testing.T, server *httptest.Server, request *db.MtrMetric, expectedCode int) {
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
func testRetrieveModel(t *testing.T, server *httptest.Server, request *db.MtrMetric, expectedCode int, expectedResponse *db.MtrMetric) {
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
func sendRequest(t *testing.T, server *httptest.Server, route string, payload interface{}) (*db.MtrMetric, int, error) {
	t.Helper()

	url := server.URL + route
	client := resty.New()
	resp, err := client.R().SetHeader("Content-Type", "application/json").SetBody(payload).Post(url)
	if err != nil {
		return nil, resp.StatusCode(), err
	}

	var metric *db.MtrMetric
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
