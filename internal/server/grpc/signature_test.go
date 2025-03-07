package grpc_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/npavlov/go-metrics-service/internal/server/grpc"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

func TestSigInterceptor(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()
	signKey := "test-secret"

	interceptor := grpc.SigInterceptor(signKey, logger)

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
	h := hmac.New(sha256.New, []byte(signKey))
	h.Write(message)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	md := metadata.Pairs("HashSHA256", expectedSignature)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor(ctx, req, nil, mockHandler)

	require.NoError(t, err)
	assert.Equal(t, "mockResponse", resp)
}
