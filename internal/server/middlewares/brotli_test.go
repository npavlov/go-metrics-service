package middlewares_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
)

func TestBrotliMiddleware(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
	})

	tests := []struct {
		name            string
		acceptEncoding  string
		expectedContent string
		expectBrotli    bool
	}{
		{
			name:            "Brotli Supported",
			acceptEncoding:  "br",
			expectedContent: "Hello, World!",
			expectBrotli:    true,
		},
		{
			name:            "Brotli Not Supported",
			acceptEncoding:  "gzip",
			expectedContent: "Hello, World!",
			expectBrotli:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			rec := httptest.NewRecorder()

			md := middlewares.ContentMiddleware("application/json")
			middleware := middlewares.BrotliMiddleware(md(handler))
			middleware.ServeHTTP(rec, req)

			if tt.expectBrotli {
				assert.Equal(t, "br", rec.Header().Get("Content-Encoding"))

				// Decompress the Brotli content here if you want to further assert the content
				br := brotli.NewReader(rec.Body)
				body, err := io.ReadAll(br)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedContent, string(body))
			} else {
				assert.NotEqual(t, "br", rec.Header().Get("Content-Encoding"))
				assert.Equal(t, tt.expectedContent, rec.Body.String())
			}
		})
	}
}
