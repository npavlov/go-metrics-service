package grpc

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingServerInterceptor logs incoming requests and responses.
func LoggingServerInterceptor(logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Start time
		start := time.Now()

		// Log the request details before handling
		logger.Info().
			Str("method", info.FullMethod).
			Interface("request", req).
			Msg("gRPC Request received")

		// Call the actual handler
		resp, err := handler(ctx, req)

		// Calculate the duration
		duration := time.Since(start)

		// Log the response details
		logEvent := logger.Info().
			Str("method", info.FullMethod).
			Dur("duration", duration)

		// Add status code and error details if there's an error
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				logEvent = logEvent.
					Int("status", int(st.Code())).
					Str("error", st.Message())
			} else {
				logEvent = logEvent.
					Int("status", int(codes.Unknown)).
					Str("error", err.Error())
			}
		} else {
			logEvent = logEvent.
				Int("status", int(codes.OK))
		}

		// Log the final message
		logEvent.Msg("gRPC Request completed")

		return resp, err
	}
}
