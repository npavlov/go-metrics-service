package grpcsender

import (
	"context"
	"net"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	au "github.com/npavlov/go-metrics-service/internal/agent/utils"
	"github.com/npavlov/go-metrics-service/pkg/crypto"
)

func MakeConnection(cfg *config.Config, logger *zerolog.Logger) *grpc.ClientConn {
	ip := au.GetLocalIP(logger)

	var encryption *crypto.Encryption

	if key := cfg.CryptoKey; key != "" {
		var err error
		if encryption, err = crypto.NewEncryption(key); err != nil {
			logger.Fatal().Err(err).Msg("failed to create encryption")
		}
	}

	interceptors := grpc.WithChainUnaryInterceptor(
		HeadersInterceptor(cfg, ip, logger),
		EncodingInterceptor(encryption, logger),
	)

	conn, err := grpc.NewClient(cfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), interceptors)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create grpc connection")
	}

	return conn
}

func MakeInMemoryConnection(cfg *config.Config, listener *bufconn.Listener, logger *zerolog.Logger) *grpc.ClientConn {
	ip := au.GetLocalIP(logger)

	var encryption *crypto.Encryption

	if key := cfg.CryptoKey; key != "" {
		var err error
		if encryption, err = crypto.NewEncryption(key); err != nil {
			logger.Fatal().Err(err).Msg("failed to create encryption")
		}
	}

	interceptors := grpc.WithChainUnaryInterceptor(
		HeadersInterceptor(cfg, ip, logger),
		EncodingInterceptor(encryption, logger),
	)

	dialer := func(_ context.Context, _ string) (net.Conn, error) {
		return listener.Dial()
	}

	// Use "passthrough" as the address since the dialer will handle the connection.
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		interceptors)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create grpc connection")
	}

	return conn
}
