package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/pkg/crypto"
)

// DecryptMiddleware - the net/http middleware function to sign http content.
func DecryptMiddleware(decryption *crypto.Decryption, log *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			// Check if the request is encoded with Gzip
			if request.Header.Get("X-Encrypted") != "true" {
				next.ServeHTTP(response, request)

				return
			}

			bodyBytes, err := io.ReadAll(request.Body)
			if err != nil {
				log.Warn().Err(err).Msg("failed to read body")
			}
			if err := request.Body.Close(); err != nil {
				log.Error().Err(err).Msg("failed to close body")
				response.WriteHeader(http.StatusInternalServerError)
			}
			data, err := decryption.Decrypt(bodyBytes)
			if err != nil {
				log.Error().Err(err).Msg("failed to decrypt body")
				response.WriteHeader(http.StatusBadRequest)
			}

			request.Body = io.NopCloser(bytes.NewBuffer(data))

			next.ServeHTTP(response, request)
		})
	}
}
