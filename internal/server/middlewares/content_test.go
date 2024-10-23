package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
)

func TestContentMiddleware(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
	})

	tests := []struct {
		name         string
		contentType  string
		expectedType string
	}{
		{
			name:         "Set Content-Type to text/plain",
			contentType:  "text/plain",
			expectedType: "text/plain",
		},
		{
			name:         "Set Content-Type to application/json",
			contentType:  "application/json",
			expectedType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			middleware := middlewares.ContentMiddleware(tt.contentType)(handler)
			middleware.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedType, rec.Header().Get("Content-Type"))
			assert.Equal(t, "Hello, World!", rec.Body.String())
		})
	}
}
