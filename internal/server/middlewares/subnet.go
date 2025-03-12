package middlewares

import (
	"net"
	"net/http"

	"github.com/rs/zerolog"
)

// SubnetMiddleware - the net/http middleware function to verify is message received from trusted subnet.
func SubnetMiddleware(subnet string, log *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if subnet == "" {
				next.ServeHTTP(response, request)
			}

			ipStr := request.Header.Get("X-Real-IP")

			if ipStr == "" {
				response.WriteHeader(http.StatusForbidden)

				return
			}

			_, trustedNet, err := net.ParseCIDR(subnet)
			if err != nil {
				log.Error().Err(err).Str("subnet", subnet).Msg("invalid subnet")
				response.WriteHeader(http.StatusInternalServerError)

				return
			}

			ip := net.ParseIP(ipStr)
			if !trustedNet.Contains(ip) {
				response.WriteHeader(http.StatusForbidden)

				return
			}

			next.ServeHTTP(response, request)
		})
	}
}
