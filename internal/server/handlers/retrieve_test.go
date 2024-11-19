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

type want struct {
	statusCode int
	result     string
}

func TestRetrieveHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request string
		data    *db.Metric
		want    want
	}{
		{
			name:    "Good value simple #1",
			request: "/value/gauge/MSpanInuse",
			data:    db.NewMetric("MSpanInuse", domain.Gauge, nil, float64Ptr(23360)),
			want: want{
				statusCode: http.StatusOK,
				result:     "23360",
			},
		},
		{
			name:    "Good value simple #2",
			request: "/value/counter/PollCount",
			data:    db.NewMetric("PollCount", domain.Counter, int64Ptr(100), nil),
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
			log := testutils.GetTLogger()
			memStorage := storage.NewMemStorage(log)
			mHandlers := handlers.NewMetricsHandler(memStorage, log)
			cfg := config.NewConfigBuilder(log).Build()
			var cRouter router.Router = router.NewCustomRouter(cfg, log)
			cRouter.SetRouter(mHandlers, nil)

			// Start the test server
			server := httptest.NewServer(cRouter.GetRouter())
			defer server.Close()

			if tt.data != nil {
				err := memStorage.Update(context.Background(), tt.data)
				if err != nil {
					t.Fatal(err)
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
