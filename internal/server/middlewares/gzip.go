package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/npavlov/go-metrics-service/internal/server/middlewares/helpers"
)

// GzipMiddleware compresses the response using Gzip if the client supports it and Brotli is not supported.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the client accepts Gzip encoding and doesn't prefer Brotli
		encoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(encoding, "gzip")
		supportsBrotli := strings.Contains(encoding, "br")
		if !supportsGzip || supportsBrotli {
			next.ServeHTTP(w, r)

			return
		}

		// Create a Gzip writer to compress the response
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length") // Can't know content length after compression

		gzWriter := gzip.NewWriter(w)
		defer func(gzWriter *gzip.Writer) {
			_ = gzWriter.Close()
		}(gzWriter)

		// Wrap the response writer
		gzipResponseWriter := &helpers.WrappedResponseWriter{ResponseWriter: w, Writer: gzWriter}
		// Pass the request to the next handler
		next.ServeHTTP(gzipResponseWriter, r)
	})
}
