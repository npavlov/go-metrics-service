package middlewares

import (
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/npavlov/go-metrics-service/internal/server/middlewares/helpers"
)

// BrotliMiddleware compresses the response using Brotli if the client supports it.
func BrotliMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the client accepts Brotli encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
			next.ServeHTTP(w, r)

			return
		}

		// Create a Brotli writer to compress the response
		w.Header().Set("Content-Encoding", "br")
		w.Header().Del("Content-Length") // Can't know content length after compression

		brWriter := brotli.NewWriter(w)
		defer func(brWriter *brotli.Writer) {
			_ = brWriter.Close()
		}(brWriter)

		// Wrap the response writer
		brResponseWriter := &helpers.WrappedResponseWriter{ResponseWriter: w, Writer: brWriter}

		// Pass the request to the next handler
		next.ServeHTTP(brResponseWriter, r)
	})
}
