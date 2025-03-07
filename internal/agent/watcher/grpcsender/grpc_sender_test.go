package grpcsender_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher/grpcsender"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

type MockMetricServiceServer struct {
	pb.UnimplementedMetricServiceServer
	mock.Mock
	cfg *config.Config
}

func NewMockMetricServiceServer(cfg *config.Config) *MockMetricServiceServer {
	return &MockMetricServiceServer{
		cfg: cfg,
	}
}

func (m *MockMetricServiceServer) SetMetrics(ctx context.Context, req *pb.SetMetricsRequest) (*pb.SetMetricsResponse, error) {
	args := m.Called(ctx, req)

	resp, ok := args.Get(0).(*pb.SetMetricsResponse)
	if !ok {
		return nil, args.Error(1)
	}

	return resp, args.Error(1)
}

func (m *MockMetricServiceServer) SetMetric(ctx context.Context, req *pb.SetMetricRequest) (*pb.SetMetricResponse, error) {
	args := m.Called(ctx, req)

	resp, ok := args.Get(0).(*pb.SetMetricResponse)
	if !ok {
		return nil, args.Error(1)
	}

	return resp, args.Error(1)
}

func setupTestGRPCServer(t *testing.T, listener *bufconn.Listener) *MockMetricServiceServer {
	t.Helper()

	server := grpc.NewServer()
	mockSvc := new(MockMetricServiceServer)
	pb.RegisterMetricServiceServer(server, mockSvc)

	go func() {
		if err := server.Serve(listener); err != nil {
			t.Errorf("server exited with error: %v", err)

			return
		}
	}()

	return mockSvc
}

func TestSendMetric(t *testing.T) {
	t.Parallel()

	listener := bufconn.Listen(1024 * 1024)

	mockSvc := setupTestGRPCServer(t, listener)
	logger := testutils.GetTLogger()

	cfg := &config.Config{
		CryptoKey: "testdata/test_public.key",
	}

	conn := grpcsender.MakeInMemoryConnection(cfg, listener, logger)
	grpcSender := grpcsender.NewGRPCSender(conn, logger)

	metric := db.NewMetric("test_metric", domain.Gauge, nil, float64Ptr(42.0))

	gMetric := utils.FromDBModelToGModel(metric)
	mockSvc.On("SetMetric", mock.Anything, mock.Anything).
		Return(&pb.SetMetricResponse{Metric: gMetric}, nil)

	resp, err := grpcSender.SendMetric(context.Background(), *metric)
	require.NoError(t, err)
	assert.Equal(t, metric.ID, resp.ID)
	assert.Equal(t, metric.MType, resp.MType)
	assert.Equal(t, metric.Value, resp.Value)
}

func TestSendMetricsBatch(t *testing.T) {
	t.Parallel()

	listener := bufconn.Listen(1024 * 1024)

	mockSvc := setupTestGRPCServer(t, listener)
	logger := testutils.GetTLogger()

	cfg := &config.Config{
		CryptoKey: "testdata/test_public.key",
	}

	conn := grpcsender.MakeInMemoryConnection(cfg, listener, logger)
	grpcSender := grpcsender.NewGRPCSender(conn, logger)

	metrics := []db.Metric{
		*db.NewMetric("test_metric", domain.Gauge, nil, float64Ptr(42.0)),
		*db.NewMetric("test_counter", domain.Counter, int64Ptr(1), nil),
	}

	gMetrics := []*pb.Metric{
		utils.FromDBModelToGModel(&metrics[0]),
		utils.FromDBModelToGModel(&metrics[1]),
	}

	mockSvc.On("SetMetrics", mock.Anything, mock.Anything).
		Return(&pb.SetMetricsResponse{Items: gMetrics}, nil)

	resp, err := grpcSender.SendMetricsBatch(context.Background(), metrics)
	require.NoError(t, err)
	assert.Len(t, resp, len(metrics))
	assert.Equal(t, metrics[0].ID, resp[0].ID)
	assert.Equal(t, metrics[1].ID, resp[1].ID)
}

func TestSendMetricsBatchNoCrypto(t *testing.T) {
	t.Parallel()

	listener := bufconn.Listen(1024 * 1024)

	mockSvc := setupTestGRPCServer(t, listener)
	logger := testutils.GetTLogger()

	cfg := &config.Config{}

	conn := grpcsender.MakeInMemoryConnection(cfg, listener, logger)
	grpcSender := grpcsender.NewGRPCSender(conn, logger)

	metrics := []db.Metric{
		*db.NewMetric("test_metric", domain.Gauge, nil, float64Ptr(42.0)),
		*db.NewMetric("test_counter", domain.Counter, int64Ptr(1), nil),
	}

	gMetrics := []*pb.Metric{
		utils.FromDBModelToGModel(&metrics[0]),
		utils.FromDBModelToGModel(&metrics[1]),
	}

	mockSvc.On("SetMetrics", mock.Anything, mock.Anything).
		Return(&pb.SetMetricsResponse{Items: gMetrics}, nil)

	resp, err := grpcSender.SendMetricsBatch(context.Background(), metrics)
	require.NoError(t, err)
	assert.Len(t, resp, len(metrics))
	assert.Equal(t, metrics[0].ID, resp[0].ID)
	assert.Equal(t, metrics[1].ID, resp[1].ID)
}
