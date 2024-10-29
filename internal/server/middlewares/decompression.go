package middlewares

import (
	"compress/gzip"
	"net/http"
)

// GzipDecompressionMiddleware decompresses the request body if it is gzip encoded.
func GzipDecompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		// Check if the request is encoded with Gzip
		if request.Header.Get("Content-Encoding") != "gzip" {
			next.ServeHTTP(response, request)

			return
		}

		// Decompress the Gzip data
		gzReader, err := gzip.NewReader(request.Body)
		if err != nil {
			http.Error(response, "Failed to decompress request body", http.StatusBadRequest)

			return
		}
		defer func(gzReader *gzip.Reader) {
			_ = gzReader.Close()
		}(gzReader)

		// Replace the request body with the decompressed data
		request.Body = gzReader
		next.ServeHTTP(response, request)
	})
}
