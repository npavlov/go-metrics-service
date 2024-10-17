package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type want struct {
	statusCode int
	result     string
}

func TestRetrieveHandler(t *testing.T) {
	var memStorage storage.Repository = storage.NewMemStorage()
	var r = chi.NewRouter()
	NewMetricsHandler(memStorage, r).SetRouter()

	server := httptest.NewServer(r)
	defer server.Close()

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
			},
			want: want{
				statusCode: http.StatusOK,
				result:     "100",
			},
		},
		{
			name:    "Non existing counter",
			request: "/value/counter/NewCounter",
			want: want{
				statusCode: http.StatusNotFound,
				result:     "",
			},
		},
		{
			name:    "Non existing gauge counter",
			request: "/value/gauge/Test",
			want: want{
				statusCode: http.StatusNotFound,
				result:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data != nil {
				switch tt.data.metricType {
				case domain.Counter:
					err := memStorage.UpdateMetric(domain.Counter, tt.data.name, tt.data.counter)
					assert.Nil(t, err)
				case domain.Gauge:
					err := memStorage.UpdateMetric(domain.Gauge, tt.data.name, tt.data.gauge)
					assert.Nil(t, err)
				default:
					t.Errorf("Invalid metric type: %s", tt.data.metricType)
				}
			}

			testRetrieveRequest(t, server, tt.request, tt.want)

		})
	}
}
func testRetrieveRequest(t *testing.T, ts *httptest.Server, route string, tt want) {
	req := resty.New().R()
	req.Method = http.MethodGet
	req.URL = ts.URL + route

	res, err := req.Send()

	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, tt.statusCode, res.StatusCode())

	if res.StatusCode() == http.StatusOK {
		assert.Equal(t, tt.result, string(res.Body()))
	}
}
