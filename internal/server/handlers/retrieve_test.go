package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type want struct {
	statusCode int
	result     string
}

func TestRetrieveHandler(t *testing.T) {
	t.Parallel()

	type metric struct {
		name       domain.MetricName
		metricType domain.MetricType
		gauge      string
		counter    string
	}

	tests := []struct {
		name    string
		request string
		data    *metric
		want    want
	}{
		{
			name:    "Good value simple #1",
			request: "/value/gauge/MSpanInuse",
			data: &metric{
				name:       "MSpanInuse",
				metricType: "gauge",
				gauge:      "23360",
				counter:    "",
			},
			want: want{
				statusCode: http.StatusOK,
				result:     "23360",
			},
		},
		{
			name:    "Good value simple #2",
			request: "/value/counter/PollCount",
			data: &metric{
				name:       "PollCount",
				metricType: "counter",
				counter:    "100",
				gauge:      "",
			},
			want: want{
				statusCode: http.StatusOK,
				result:     "100",
			},
		},
		{
			name:    "Non existing counter",
			request: "/value/counter/NewCounter",
			data:    nil,
			want: want{
				statusCode: http.StatusNotFound,
				result:     "",
			},
		},
		{
			name:    "Non existing gauge counter",
			request: "/value/gauge/Test",
			data:    nil,
			want: want{
				statusCode: http.StatusNotFound,
				result:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Initialize storage and router
			var memStorage storage.Repository = storage.NewMemStorage()
			r := chi.NewRouter()
			handlers.NewMetricsHandler(memStorage, r).SetRouter()

			// Start the test server
			server := httptest.NewServer(r)
			defer server.Close()

			if tt.data != nil {
				switch tt.data.metricType {
				case domain.Counter:
					err := memStorage.UpdateMetric(domain.Counter, tt.data.name, tt.data.counter)
					require.NoError(t, err)
				case domain.Gauge:
					err := memStorage.UpdateMetric(domain.Gauge, tt.data.name, tt.data.gauge)
					require.NoError(t, err)
				default:
					t.Errorf("Invalid metric type: %s", tt.data.metricType)
				}
			}

			testRetrieveRequest(t, server, tt.request, tt.want)
		})
	}
}

func testRetrieveRequest(t *testing.T, ts *httptest.Server, route string, tt want) {
	t.Helper()

	req := resty.New().R()
	req.Method = http.MethodGet
	req.URL = ts.URL + route

	res, err := req.Send()

	require.NoError(t, err, "error making HTTP request")
	assert.Equal(t, tt.statusCode, res.StatusCode())

	if res.StatusCode() == http.StatusOK {
		assert.Equal(t, tt.result, string(res.Body()))
	}
}
