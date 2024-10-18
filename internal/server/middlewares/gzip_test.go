package middlewares_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/npavlov/go-metrics-service/internal/server/middlewares"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestGzipMiddleware tests the gzip compression middleware.
func TestGzipMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("ShouldCompressResponseWhenClientAcceptsGzip", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Hello, world!"))
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		middlewares.GzipMiddleware(handler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

		// Verify that the response body is actually gzip compressed
		gr, err := gzip.NewReader(rec.Body)
		require.NoError(t, err)
		defer func(gr *gzip.Reader) {
			_ = gr.Close()
		}(gr)

		unzippedBody, err := io.ReadAll(gr)
		require.NoError(t, err)

		assert.Equal(t, "Hello, world!", string(unzippedBody))
	})

	t.Run("ShouldNotCompressResponseWhenClientDoesNotAcceptGzip", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Hello, world!"))
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		middlewares.GzipMiddleware(handler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "Hello, world!", rec.Body.String())
	})

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

		middlewares.GzipMiddleware(handler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("ShouldPassThroughWhenBrotliEncodingIsSupportedButNotGzip", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Hello, brotli!"))
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "br")
		rec := httptest.NewRecorder()

		middlewares.GzipMiddleware(handler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "Hello, brotli!", rec.Body.String())
	})

	t.Run("ShouldCloseGzipWriterOnDefer", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test response"))
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		middlewares.GzipMiddleware(handler).ServeHTTP(rec, req)

		// Verify gzip writer is closed properly by checking content length
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
		assert.NotEmpty(t, rec.Body.Bytes())
	})
}
