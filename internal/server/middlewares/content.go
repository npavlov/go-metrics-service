package middlewares

import (
	"net/http"
)

func ContentMiddleware(contentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)

			next.ServeHTTP(w, r)
		})
	}
}
