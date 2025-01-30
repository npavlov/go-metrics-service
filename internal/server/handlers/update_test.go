package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/config"

	"github.com/npavlov/go-metrics-service/internal/server/db"

	"github.com/npavlov/go-metrics-service/internal/server/router"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestUpdateHandler(t *testing.T) {
	t.Parallel()

	type metric struct {
		name       domain.MetricName
		metricType domain.MetricType
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
		initial *metric
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
					counter:    0,
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
					gauge:      0,
				},
			},
		},
		{
			name:    "Update counter if found",
			request: "/update/counter/PollCount/2",
			want: want{
				statusCode: http.StatusOK,
				result: &metric{
					name:       "PollCount",
					metricType: "counter",
					counter:    3,
					gauge:      0,
				},
			},
			initial: &metric{
				name:       "PollCount",
				metricType: "counter",
				counter:    1,
				gauge:      0,
			},
		},
		{
			name:    "Update existing gauge metric",
			request: "/update/gauge/ExistingGauge/250",
			initial: &metric{
				name:       "ExistingGauge",
				metricType: domain.Gauge,
				gauge:      100.0,
			},
			want: want{
				statusCode: http.StatusOK,
				result: &metric{
					name:       "ExistingGauge",
					metricType: domain.Gauge,
					gauge:      250.0,
				},
			},
		},
		{
			name:    "Update existing counter metric",
			request: "/update/counter/ExistingCounter/15",
			initial: &metric{
				name:       "ExistingCounter",
				metricType: domain.Counter,
				counter:    10,
			},
			want: want{
				statusCode: http.StatusOK,
				result: &metric{
					name:       "ExistingCounter",
					metricType: domain.Counter,
					counter:    25, // 10 + 15 as it should increment
				},
			},
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

			if tt.initial != nil {
				mod := db.NewMetric(tt.initial.name, tt.initial.metricType, int64Ptr(tt.initial.counter), float64Ptr(tt.initial.gauge))

				_ = memStorage.Create(context.Background(), mod)
			}

			// Start the test server
			server := httptest.NewServer(cRouter.GetRouter())
			defer server.Close()

			testUpdateRequest(t, server, tt.request, tt.want.statusCode)

			if tt.want.result != nil {
				metric, exist := memStorage.Get(context.Background(), tt.want.result.name)
				assert.True(t, exist)

				switch metric.MType {
				case domain.Gauge:
					assert.InDelta(t, tt.want.result.gauge, *metric.Value, 0.0001)

				case domain.Counter:
					assert.Equal(t, tt.want.result.counter, *metric.Delta)
				}
			}
		})
	}
}

func testUpdateRequest(t *testing.T, ts *httptest.Server, route string, statusCode int) {
	t.Helper()

	req := resty.New().R()
	req.Method = http.MethodPost
	req.URL = ts.URL + route

	res, err := req.Send()

	require.NoError(t, err, "error making HTTP request")
	assert.Equal(t, statusCode, res.StatusCode())
}
