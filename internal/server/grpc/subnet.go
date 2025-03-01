package grpc

import (
	"context"
	"net"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// SubnetInterceptor verifies if the request is from a trusted subnet.
func SubnetInterceptor(subnet string, log *zerolog.Logger) grpc.UnaryServerInterceptor {
	_, trustedNet, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid subnet configuration")
	}

	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Error().Msg("missing metadata in request")

			return nil, errors.New("missing metadata")
		}

		ipStr := md.Get("X-Real-IP")
		if len(ipStr) == 0 {
			log.Error().Msg("X-Real-IP header is missing")

			return nil, errors.New("X-Real-IP header is required")
		}

		ip := net.ParseIP(strings.TrimSpace(ipStr[0]))
		if ip == nil {
			log.Error().Str("ip", ipStr[0]).Msg("invalid IP address format")

			return nil, errors.New("invalid IP address format")
		}

		if !trustedNet.Contains(ip) {
			log.Warn().Str("ip", ip.String()).Msg("unauthorized access attempt")

			return nil, errors.New("unauthorized access")
		}

		return handler(ctx, req)
	}
}
