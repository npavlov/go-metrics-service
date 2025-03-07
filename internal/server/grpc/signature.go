package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/npavlov/go-metrics-service/internal/utils"
)

// SigInterceptor checks if the request signature is valid.
func SigInterceptor(signKey string, log *zerolog.Logger) grpc.UnaryServerInterceptor {
	if signKey == "" {
		log.Fatal().Msg("signKey must be provided for signature verification")
	}

	hPool := &sync.Pool{
		New: func() interface{} {
			return hmac.New(sha256.New, []byte(signKey))
		},
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

		signatures := md.Get("HashSHA256")
		if len(signatures) == 0 {
			log.Error().Msg("missing signature header")

			return nil, errors.New("missing HashSHA256 header")
		}

		expectedSignature := signatures[0]

		// Serialize the request payload
		payload, err := utils.MarshalProtoMessage(req)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal request")

			return nil, errors.New("failed to marshal request")
		}

		// Compute HMAC signature
		hmacWriter, ok := hPool.Get().(hash.Hash)
		if !ok {
			log.Error().Msg("failed to get HMAC instance")

			return nil, errors.New("internal server error")
		}
		defer hPool.Put(hmacWriter)

		hmacWriter.Reset()
		hmacWriter.Write(payload)
		computedSignature := hex.EncodeToString(hmacWriter.Sum(nil))

		// Compare the expected vs computed signature
		if computedSignature != expectedSignature {
			return nil, errors.New("invalid signature")
		}

		return handler(ctx, req)
	}
}
