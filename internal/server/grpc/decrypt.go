package grpc

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

// DecryptInterceptor checks if the request signature is valid.
//
//nolint:cyclop
func DecryptInterceptor(decryption *crypto.Decryption, log *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if decryption == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Error().Msg("missing metadata in request")

			return nil, errors.New("missing metadata")
		}

		encryptionMds := md.Get("x-encrypted")
		if len(encryptionMds) == 0 || encryptionMds[0] != "true" {
			log.Error().Msg("missing or invalid encryption header")

			return nil, errors.New("missing or invalid encryption header")
		}

		// Handle different request types using a type switch
		switch request := req.(type) {
		case *pb.MetricRequest:
			decrypt, err := decryption.Decrypt(request.GetEncryptedMessage())
			if err != nil {
				return nil, errors.Wrap(err, "failed to decrypt request")
			}
			if err := proto.Unmarshal(decrypt, request); err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal request")
			}

			return handler(ctx, request)

		case *pb.MetricsRequest:
			decrypt, err := decryption.Decrypt(request.GetEncryptedMessage())
			if err != nil {
				return nil, errors.Wrap(err, "failed to decrypt request")
			}
			if err := proto.Unmarshal(decrypt, request); err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal request")
			}

			return handler(ctx, request)

		default:
			log.Error().Msg("invalid request type")

			return nil, errors.New("invalid request type")
		}
	}
}
