package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"

	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/router"
)

func TestNewCustomRouter(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	logger := zerolog.Nop()
	r := router.NewCustomRouter(cfg, &logger)
	assert.NotNil(t, r)
	assert.NotNil(t, r.GetRouter())
}

func TestNewCustomRouterWithCryptoKey(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{CryptoKey: "testdata/test_private.key"}
	logger := testutils.GetTLogger()
	customRouter := router.NewCustomRouter(cfg, logger)
	mh := &handlers.MetricHandler{}
	hh := &handlers.HealthHandler{}
	assert.NotNil(t, customRouter)
	customRouter.SetRouter(mh, hh)
	r := customRouter.GetRouter()
	assert.Len(t, r.Middlewares(), 8)
}

func TestNewCustomRouterWithBrokenCryptoKey(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{CryptoKey: ""}
	logger := testutils.GetTLogger()
	customRouter := router.NewCustomRouter(cfg, logger)
	mh := &handlers.MetricHandler{}
	hh := &handlers.HealthHandler{}
	assert.NotNil(t, customRouter)
	customRouter.SetRouter(mh, hh)
	mux := customRouter.GetRouter()
	assert.Len(t, mux.Middlewares(), 7)
}

func TestSetRouter(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	logger := zerolog.Nop()
	customRouter := router.NewCustomRouter(cfg, &logger)

	// Create mock handlers
	mh := &handlers.MetricHandler{}
	hh := &handlers.HealthHandler{}

	customRouter.SetRouter(mh, hh)
	mux := customRouter.GetRouter()
	require.NotNil(t, mux)
}

func TestRoutes(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	logger := zerolog.Nop()
	customRouter := router.NewCustomRouter(cfg, &logger)
	mh := &handlers.MetricHandler{}
	hh := &handlers.HealthHandler{}

	customRouter.SetRouter(mh, hh)
	mux := customRouter.GetRouter()

	tests := []struct {
		name       string
		method     string
		url        string
		statusCode int
	}{
		{"Root Route", "GET", "/", http.StatusOK},
		{"Ping Route", "GET", "/ping", http.StatusOK},
		{"Update Metric", "POST", "/update/gauge/cpu/100", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}
