package middlewares

import (
	"compress/gzip"
	"net/http"
)

// GzipDecompressionMiddleware decompresses the request body if it is gzip encoded.
func GzipDecompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is encoded with Gzip
		if r.Header.Get("Content-Encoding") != "gzip" {
			next.ServeHTTP(w, r)

			return
		}

		// Decompress the Gzip data
		gzReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Failed to decompress request body", http.StatusBadRequest)

			return
		}
		defer func(gzReader *gzip.Reader) {
			_ = gzReader.Close()
		}(gzReader)

		// Replace the request body with the decompressed data
		r.Body = gzReader
		next.ServeHTTP(w, r)
	})
}
