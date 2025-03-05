package grpcsender_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher/grpcsender"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
	"github.com/npavlov/go-metrics-service/pkg/crypto"
)

// MockUnaryInvoker is a mock implementation of the grpc.UnaryInvoker function.
type MockUnaryInvoker struct {
	mock.Mock
}

func (m *MockUnaryInvoker) Invoke(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
	args := m.Called(ctx, method, req, reply, cc, opts)

	return args.Error(0)
}

func TestEncodingInterceptor_NoEncryption(t *testing.T) {
	t.Parallel()

	encryption := (*crypto.Encryption)(nil)
	interceptor := grpcsender.EncodingInterceptor(encryption, nil)

	ctx := context.Background()
	method := pb.MetricService_SetMetric_FullMethodName
	req := &pb.SetMetricRequest{Metric: &pb.Metric{Id: "test", Mtype: pb.Metric_TYPE_GAUGE, Value: float64Ptr(1)}}
	reply := &pb.SetMetricResponse{}
	mockConn := &grpc.ClientConn{}

	invoker := &MockUnaryInvoker{}
	invoker.On("Invoke", ctx, method, req, reply, mockConn, mock.Anything).Return(nil)

	err := interceptor(ctx, method, req, reply, mockConn, invoker.Invoke)
	require.NoError(t, err)
	invoker.AssertCalled(t, "Invoke", ctx, method, req, reply, mockConn, mock.Anything)
}

func TestEncodingInterceptor_WithEncryption(t *testing.T) {
	t.Parallel()

	encryption, _ := crypto.NewEncryption("testdata/test_public.key")
	logger := testutils.GetTLogger()
	interceptor := grpcsender.EncodingInterceptor(encryption, logger)

	ctx := context.Background()
	method := pb.MetricService_SetMetric_FullMethodName
	req := &pb.SetMetricRequest{Metric: &pb.Metric{Id: "test", Mtype: pb.Metric_TYPE_GAUGE, Value: float64Ptr(1)}}
	reply := &pb.SetMetricResponse{}
	mockConn := &grpc.ClientConn{}

	invoker := &MockUnaryInvoker{}
	invoker.On("Invoke", mock.Anything, method, mock.Anything, reply, mockConn, mock.Anything).Return(nil)

	marshaledData, err := proto.Marshal(req)
	require.NoError(t, err)

	err = interceptor(ctx, method, req, reply, mockConn, invoker.Invoke)
	require.NoError(t, err)

	invoker.AssertCalled(t, "Invoke", mock.Anything, method, mock.Anything, reply, mockConn, mock.Anything)
	assert.Nil(t, req.GetMetric())

	decryption, err := crypto.NewDecryption("testdata/test_private.key")
	require.NoError(t, err)

	decrypted, err := decryption.Decrypt(req.GetEncryptedMessage())
	require.NoError(t, err)

	assert.Equal(t, marshaledData, decrypted)
}

func TestEncodingInterceptor_ErrorHandling(t *testing.T) {
	t.Parallel()

	encryption, _ := crypto.NewEncryption("testdata/test_public.key")
	logger := testutils.GetTLogger()
	interceptor := grpcsender.EncodingInterceptor(encryption, logger)

	ctx := context.Background()
	method := pb.MetricService_SetMetric_FullMethodName
	reply := &pb.SetMetricResponse{}
	mockConn := &grpc.ClientConn{}
	invoker := &MockUnaryInvoker{}
	invoker.On("Invoke", mock.Anything, method, mock.Anything, reply, mockConn, mock.Anything).Return(nil)

	err := interceptor(ctx, method, nil, reply, mockConn, invoker.Invoke)

	invoker.AssertNotCalled(t, "Invoke", mock.Anything, method, mock.Anything, reply, mockConn, mock.Anything)
	require.Error(t, err)
	assert.Equal(t, "failed to convert request to proto.Message", err.Error())
}

func float64Ptr(f float64) *float64 {
	return &f
}
