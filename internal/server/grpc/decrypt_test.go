package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/npavlov/go-metrics-service/internal/server/grpc"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
	"github.com/npavlov/go-metrics-service/internal/utils"
	"github.com/npavlov/go-metrics-service/pkg/crypto"
)

// Test DecryptInterceptor.
func TestDecryptInterceptor(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()
	encryption, _ := crypto.NewEncryption("testdata/test_public.key")
	decryption, _ := crypto.NewDecryption("testdata/test_private.key")

	interceptor := grpc.DecryptInterceptor(decryption, logger)

	// Create request
	req := &pb.SetMetricRequest{
		Metric: &pb.Metric{
			Id:    "test_metric",
			Mtype: pb.Metric_TYPE_COUNTER,
			Delta: int64Ptr(100),
		},
	}
	message, err := utils.MarshalProtoMessage(req)
	require.NoError(t, err)
	encryptedData, err := encryption.Encrypt(message)
	require.NoError(t, err)
	newReq := &pb.SetMetricRequest{
		EncryptedMessage: encryptedData,
	}

	md := metadata.Pairs("x-encrypted", "true")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor(ctx, newReq, nil, mockHandler)

	require.NoError(t, err)
	assert.Equal(t, "mockResponse", resp)
	assert.Equal(t, newReq.GetMetric().GetId(), req.GetMetric().GetId())
}

// Test DecryptInterceptor.
func TestDecryptInterceptorSetMetrics(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()
	encryption, _ := crypto.NewEncryption("testdata/test_public.key")
	decryption, _ := crypto.NewDecryption("testdata/test_private.key")

	interceptor := grpc.DecryptInterceptor(decryption, logger)

	// Create request
	req := &pb.SetMetricsRequest{
		Items: []*pb.Metric{
			{
				Id:    "test_metric",
				Mtype: pb.Metric_TYPE_COUNTER,
				Delta: int64Ptr(100),
			},
			{
				Id:    "test_metric2",
				Mtype: pb.Metric_TYPE_GAUGE,
				Value: float64Ptr(42.2),
			},
		},
	}
	message, err := utils.MarshalProtoMessage(req)
	require.NoError(t, err)
	encryptedData, err := encryption.Encrypt(message)
	require.NoError(t, err)
	newReq := &pb.SetMetricsRequest{
		EncryptedMessage: encryptedData,
	}

	md := metadata.Pairs("x-encrypted", "true")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor(ctx, newReq, nil, mockHandler)

	require.NoError(t, err)
	assert.Equal(t, "mockResponse", resp)
	assert.Len(t, newReq.GetItems(), len(req.GetItems()))
}

// Helper function to create float64 pointer.
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create int64 pointer.
func int64Ptr(i int64) *int64 {
	return &i
}
