package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestUpdateRetrieveModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		update           *db.Metric
		retrieve         *db.Metric
		expectedCode     int
		expectedResponse *db.Metric
	}{
		{
			name:             "Successful Gauge Retrieval",
			update:           db.NewMetric("Alloc", domain.Gauge, nil, float64Ptr(100.0)),
			retrieve:         db.NewMetric("Alloc", domain.Gauge, nil, nil),
			expectedCode:     http.StatusOK,
			expectedResponse: db.NewMetric("Alloc", domain.Gauge, nil, float64Ptr(100.0)),
		},
		{
			name:             "Successful Counter Retrieval",
			update:           db.NewMetric("RequestCount", domain.Counter, int64Ptr(42), nil),
			retrieve:         db.NewMetric("RequestCount", domain.Counter, nil, nil),
			expectedCode:     http.StatusOK,
			expectedResponse: db.NewMetric("RequestCount", domain.Counter, int64Ptr(42), nil),
		},
		{
			name:         "Invalid JSON Payload on Update",
			update:       nil,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Retrieve Non-Existent Metric",
			retrieve:     db.NewMetric("non_existing", domain.Counter, nil, nil),
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
			cfg := config.NewConfigBuilder(log).Build()
			var cRouter router.Router = router.NewCustomRouter(cfg, log)
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
func testUpdateModel(t *testing.T, server *httptest.Server, request *db.Metric, expectedCode int) {
	t.Helper()

	json := jsoniter.ConfigCompatibleWithStandardLibrary

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
func testRetrieveModel(t *testing.T, server *httptest.Server, request *db.Metric, expectedCode int, expectedResponse *db.Metric) {
	t.Helper()

	json := jsoniter.ConfigCompatibleWithStandardLibrary
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
func sendRequest(t *testing.T, server *httptest.Server, route string, payload interface{}) (*db.Metric, int, error) {
	t.Helper()

	json := jsoniter.ConfigCompatibleWithStandardLibrary

	url := server.URL + route
	client := resty.New()
	resp, err := client.R().SetHeader("Content-Type", "application/json").SetBody(payload).Post(url)
	if err != nil {
		return nil, resp.StatusCode(), err
	}

	var metric *db.Metric
	if err := json.Unmarshal(resp.Body(), &metric); err != nil {
		return nil, resp.StatusCode(), err
	}

	return metric, resp.StatusCode(), nil
}

// BenchmarkRetrieveModel - Benchmark for RetrieveModel.
func BenchmarkRetrieveModel(b *testing.B) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	// Initialize storage and router
	log := testutils.GetTLogger()
	memStorage := storage.NewMemStorage(log)
	mHandlers := handlers.NewMetricsHandler(memStorage, log)

	updateModel := db.NewMetric("Alloc", domain.Gauge, nil, float64Ptr(100.0))

	_ = memStorage.Create(context.Background(), updateModel)

	retrieveModel := db.NewMetric("Alloc", domain.Gauge, nil, nil)
	// Prepare a valid JSON request body
	requestBody, _ := json.Marshal(retrieveModel)

	// Benchmark loop
	for range b.N {
		// Create a new HTTP request
		request := httptest.NewRequest(http.MethodPost, "/retrieve", bytes.NewReader(requestBody))

		// Create a new ResponseRecorder to capture the response
		response := httptest.NewRecorder()

		// Call the handler
		mHandlers.RetrieveModel(response, request)
	}
}

// Helper function to create float64 pointer.
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create int64 pointer.
func int64Ptr(i int64) *int64 {
	return &i
}
