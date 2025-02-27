package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestSubnetMiddleware(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()

	// Create a handler that just returns OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	tests := []struct {
		name           string
		ipHeader       string
		expectedStatus int
		subnet         string
	}{
		{"Allowed IP", "192.168.1.100", http.StatusOK, "192.168.1.0/24"},
		{"Forbidden IP", "10.0.0.1", http.StatusForbidden, "192.168.1.0/24"},
		{"Missing IP Header", "", http.StatusForbidden, "192.168.1.0/24"},
		{"Broken subnet", "192.168.1.100", http.StatusInternalServerError, "111"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			middleware := middlewares.SubnetMiddleware(tt.subnet, logger)

			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			if tt.ipHeader != "" {
				req.Header.Set("X-Real-IP", tt.ipHeader)
			}

			recorder := httptest.NewRecorder()
			middleware(handler).ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}
