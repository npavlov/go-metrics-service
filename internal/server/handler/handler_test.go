package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	types "github.com/npavlov/go-metrics-service/internal/agent/metrictypes"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateHandler(t *testing.T) {
	var memStorage storage.Repository = storage.NewMemStorage()

	handler := GetUpdateHandler(memStorage)

	type metric struct {
		name       types.MetricName
		metricType types.MetricType
		gauge      float64
		counter    int64
	}

	type want struct {
		statusCode int
		result     *metric
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Without metric name",
			request: "/update/gauge/",
			want: want{
				statusCode: http.StatusNotFound,
				result:     nil,
			},
		},
		{
			name:    "Without metric value",
			request: "/update/gauge/MSpanInuse/",
			want: want{
				statusCode: http.StatusNotFound,
				result:     nil,
			},
		},
		{
			name:    "With bad value",
			request: "/update/gauge/MSpanInuse/text",
			want: want{
				statusCode: http.StatusBadRequest,
				result:     nil,
			},
		},
		{
			name:    "With bad value counter",
			request: "/update/counter/MSpanInuse/text",
			want: want{
				statusCode: http.StatusBadRequest,
				result:     nil,
			},
		},
		{
			name:    "With unknown metric type",
			request: "/update/unknown/MSpanInuse/2324.43",
			want: want{
				statusCode: http.StatusBadRequest,
				result:     nil,
			},
		},
		{
			name:    "Good value simple #1",
			request: "/update/gauge/MSpanInuse/23360.000000",
			want: want{
				statusCode: http.StatusOK,
				result: &metric{
					name:       "MSpanInuse",
					metricType: "gauge",
					gauge:      23360,
				},
			},
		},
		{
			name:    "Good value simple #2",
			request: "/update/counter/PollCount/100",
			want: want{
				statusCode: http.StatusOK,
				result: &metric{
					name:       "PollCount",
					metricType: "counter",
					counter:    100,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()

			router.SetRoutes(r, handler)

			server := httptest.NewServer(r)
			defer server.Close()

			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = server.URL + tt.request

			res, err := req.Send()

			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.want.statusCode, res.StatusCode())

			if tt.want.result != nil {
				switch tt.want.result.metricType {
				case types.Gauge:
					value, exist := memStorage.GetGauge(tt.want.result.name)
					assert.True(t, exist)
					assert.Equal(t, tt.want.result.gauge, value)
				case types.Counter:
					value, exist := memStorage.GetCounter(tt.want.result.name)
					assert.True(t, exist)
					assert.Equal(t, tt.want.result.counter, value)
				default:
					t.Errorf("Unexpected metric type: %v", tt.want.result.metricType)
				}
			}

		})
	}
}
