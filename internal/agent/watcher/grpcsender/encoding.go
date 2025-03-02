package grpcsender

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/npavlov/go-metrics-service/pkg/crypto"
	pb "github.com/npavlov/go-metrics-service/proto/v1"
)

// EncodingInterceptor crypts outgoing message.
func EncodingInterceptor(encryption *crypto.Encryption, logger *zerolog.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string, req,
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if encryption == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		// Convert request to bytes
		reqProto, ok := req.(proto.Message)
		if !ok {
			logger.Error().Msg("Failed to convert request to proto.Message")

			return invoker(ctx, method, req, reply, cc, opts...)
		}

		reqBytes, err := proto.Marshal(reqProto)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to marshal request")

			return errors.Wrap(err, "Failed to marshal request")
		}

		encryptedPayload, err := encryption.Encrypt(reqBytes)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to encrypt payload")

			return errors.Wrap(err, "failed to encrypt payload")
		}

		// Define the metadata key-value pairs
		kv := []string{"x-encrypted", "true"}

		// Append the metadata to the context
		newCtx := metadata.AppendToOutgoingContext(ctx, kv...)

		if request, ok := req.(*pb.MetricRequest); ok {
			request.EncryptedMessage = encryptedPayload
			request.Metric = nil

			return invoker(newCtx, method, request, reply, cc, opts...)
		}

		if request, ok := req.(*pb.MetricsRequest); ok {
			request.EncryptedMessage = encryptedPayload
			request.Items = nil

			return invoker(newCtx, method, request, reply, cc, opts...)
		}

		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}
