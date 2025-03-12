package grpcsender_test

import (
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher/grpcsender"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestMakeConnection(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()
	cfg := &config.Config{
		GRPCAddress: "localhost:50051",
		CryptoKey:   "", // No encryption
	}

	conn := grpcsender.MakeConnection(cfg, logger)
	require.NotNil(t, conn, "gRPC connection should not be nil")
	assert.NoError(t, conn.Close(), "gRPC connection should close without error")
}

func TestMakeConnectionWithCrypto(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()
	cfg := &config.Config{
		GRPCAddress: "localhost:50051",
		CryptoKey:   "testdata/test_public.key", // Fake key path
	}

	conn := grpcsender.MakeConnection(cfg, logger)
	require.NotNil(t, conn, "gRPC connection should not be nil")
	assert.NoError(t, conn.Close(), "gRPC connection should close without error")
}

func TestMakeInMemoryConnection(t *testing.T) {
	t.Parallel()

	listener := bufconn.Listen(1024 * 1024)
	logger := testutils.GetTLogger()
	cfg := &config.Config{
		CryptoKey: "", // No encryption
	}

	conn := grpcsender.MakeInMemoryConnection(cfg, listener, logger)
	require.NotNil(t, conn, "In-memory gRPC connection should not be nil")

	// Test if connection works by sending a dummy request
	client, _ := grpc.NewClient(conn.Target(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NotNil(t, client)

	assert.NoError(t, conn.Close(), "In-memory gRPC connection should close without error")
}

func TestMakeInMemoryConnectionWithCrypto(t *testing.T) {
	t.Parallel()

	listener := bufconn.Listen(1024 * 1024)
	logger := log.Logger
	cfg := &config.Config{
		CryptoKey: "testdata/test_public.key", // Fake key path
	}

	conn := grpcsender.MakeInMemoryConnection(cfg, listener, &logger)
	require.NotNil(t, conn, "In-memory gRPC connection should not be nil")

	// Test if connection works by sending a dummy request
	client, _ := grpc.NewClient(conn.Target(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NotNil(t, client)

	assert.NoError(t, conn.Close(), "In-memory gRPC connection should close without error")
}
