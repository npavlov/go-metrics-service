package middlewares_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
)

// Helper function to gzip compress data.
func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(data)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func TestGzipDecompressionMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("ShouldHandleGzipRequestBody", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, "compressed body", string(body))
			w.WriteHeader(http.StatusOK)
		})

		compressedBody, err := gzipCompress([]byte("compressed body"))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(compressedBody))
		req.Header.Set("Content-Encoding", "gzip")
		rec := httptest.NewRecorder()

		middlewares.GzipDecompressionMiddleware(handler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
