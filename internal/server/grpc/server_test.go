package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/grpc"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

// Test SetMetric.
func TestSetMetric(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()
	signKey := "test-secret"
	cfg := &config.Config{
		Key: signKey,
	}
	memStorage := storage.NewMemStorage(logger)
	server := grpc.NewGRPCServer(memStorage, cfg, logger)

	req := &pb.SetMetricRequest{
		Metric: &pb.Metric{
			Id:    "test_metric",
			Mtype: pb.Metric_TYPE_COUNTER,
			Delta: int64Ptr(100),
		},
	}

	resp, err := server.SetMetric(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp.GetStatus())
	assert.Equal(t, req.GetMetric(), resp.GetMetric())
}

// Test SetMetrics.
func TestSetMetrics(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()
	signKey := "test-new-secret"
	cfg := &config.Config{
		Key: signKey,
	}
	memStorage := storage.NewMemStorage(logger)
	server := grpc.NewGRPCServer(memStorage, cfg, logger)

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

	resp, err := server.SetMetrics(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp.GetStatus())
	assert.Len(t, req.GetItems(), len(resp.GetItems()))

	req2 := &pb.SetMetricRequest{
		Metric: &pb.Metric{
			Id:    "test_metric",
			Mtype: pb.Metric_TYPE_COUNTER,
			Delta: int64Ptr(100),
		},
	}

	resp2, err := server.SetMetric(context.Background(), req2)
	require.NoError(t, err)
	assert.Equal(t, int64(200), resp2.GetMetric().GetDelta())
}
