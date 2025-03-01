package grpc

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// LoggingServerInterceptor logs incoming requests and responses.
func LoggingServerInterceptor(logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Log the encrypted request (before decryption)
		logger.Info().Interface("info", info).Msg("Request received")

		// Call the actual handler
		resp, err := handler(ctx, req)

		return resp, err
	}
}
