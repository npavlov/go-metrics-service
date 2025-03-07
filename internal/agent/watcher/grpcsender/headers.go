package grpcsender

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

// HeadersInterceptor adds X-Real-IP and HashSHA256 metadata.
func HeadersInterceptor(cfg *config.Config, ip string, logger *zerolog.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string, req,
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		md := metadata.New(map[string]string{
			"X-Real-IP": ip,
		})

		// Serialize request and calculate hash if key is set
		if cfg.Key != "" {
			payload, err := utils.MarshalProtoMessage(req)
			if err != nil {
				logger.Error().Err(err).Msg("failed to marshal request for hashing")
			} else {
				hash := utils.CalculateHash(cfg.Key, payload)
				md.Append("HashSHA256", hash)
			}
		}

		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
