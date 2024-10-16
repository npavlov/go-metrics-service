package middlewares

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := logger.Get()
		// Start time
		start := time.Now()

		// Wrap the response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			// Log the request details
			l.Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Str("remote", r.RemoteAddr).
				Dur("duration", time.Since(start)).
				Msg("HTTP Request")
		}()

		// Call the next handler in the chain
		next.ServeHTTP(ww, r)
	})
}
