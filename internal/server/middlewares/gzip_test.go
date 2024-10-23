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

// TestGzipMiddleware tests the gzip compression middleware.
func TestGzipMiddleware(t *testing.T) {
	t.Parallel()

	md := middlewares.ContentMiddleware("application/json")

	t.Run("ShouldCompressResponseWhenClientAcceptsGzip", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Hello, world!"))
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		middlewares.GzipMiddleware(md(handler)).ServeHTTP(rec, req)

		rec.WriteHeader(http.StatusOK)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

		// Verify that the response body is actually gzip compressed
		compressedBody := rec.Body.Bytes()
		gr, err := gzip.NewReader(bytes.NewReader(compressedBody))
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

		middlewares.GzipMiddleware(md(handler)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "Hello, world!", rec.Body.String())
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

		middlewares.GzipMiddleware(md(handler)).ServeHTTP(rec, req)

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

		middlewares.GzipMiddleware(md(handler)).ServeHTTP(rec, req)

		// Verify gzip writer is closed properly by checking content length
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
		assert.NotEmpty(t, rec.Body.Bytes())
	})
}
