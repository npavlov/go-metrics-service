package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/npavlov/go-metrics-service/internal/server/middlewares/helpers"
)

// GzipMiddleware compresses the response using Gzip if the client supports it and Brotli is not supported.
func GzipMiddleware(next http.Handler) http.Handler {
	gzipPool := &sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(io.Discard)
		},
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		// Check if the client accepts Gzip encoding and doesn't prefer Brotli
		encoding := request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(encoding, "gzip")
		supportsBrotli := strings.Contains(encoding, "br")
		if !supportsGzip || supportsBrotli {
			next.ServeHTTP(response, request)

			return
		}

		// Create a Gzip writer to compress the response
		response.Header().Set("Content-Encoding", "gzip")
		response.Header().Del("Content-Length") // Can't know content length after compression

		gzWriter, ok := gzipPool.Get().(*gzip.Writer)
		if !ok {
			response.WriteHeader(http.StatusInternalServerError)

			return
		}
		defer gzipPool.Put(gzWriter)

		gzWriter.Reset(response)
		defer func(gzWriter *gzip.Writer) {
			_ = gzWriter.Close()
		}(gzWriter)

		// Wrap the response writer
		gzipResponseWriter := &helpers.WrappedResponseWriter{ResponseWriter: response, Writer: gzWriter}
		// Pass the request to the next handler
		next.ServeHTTP(gzipResponseWriter, request)
	})
}
