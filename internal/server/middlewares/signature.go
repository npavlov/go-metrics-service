package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
	"sync"

	"github.com/rs/zerolog"
)

// SignatureMiddleware - the net/http middleware function to sign http content.
func SignatureMiddleware(signKey string, log *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hPool := &sync.Pool{
			New: func() interface{} {
				return hmac.New(sha256.New, []byte(signKey))
			},
		}

		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			hashSum := request.Header.Get("HashSHA256")
			if request.Method == http.MethodPost && hashSum != "" && signKey != "" {
				bodyBytes, err := io.ReadAll(request.Body)
				if err != nil {
					log.Warn().Msg("failed to read request body")
				}
				if err := request.Body.Close(); err != nil {
					response.WriteHeader(http.StatusBadRequest)

					return
				}
				request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				h := hPool.Get().(hash.Hash)
				h.Reset()
				h.Write(bodyBytes)
				signature := hex.EncodeToString(h.Sum(nil))

				defer hPool.Put(h)

				response.Header().Add("HashSHA256", signature)

				if signature != hashSum {
					log.Error().Msg("invalid signature")

					response.WriteHeader(http.StatusBadRequest)

					return
				}
			}

			next.ServeHTTP(response, request)
		})
	}
}
