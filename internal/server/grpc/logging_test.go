package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcserver "github.com/npavlov/go-metrics-service/internal/server/grpc"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

// Mock handler.
func mockHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return "mockResponse", nil
}

// Test LoggingServerInterceptor.
func TestLoggingServerInterceptor(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()

	interceptor := grpcserver.LoggingServerInterceptor(logger)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}

	resp, err := interceptor(context.Background(), "testRequest", info, mockHandler)

	require.NoError(t, err)
	assert.Equal(t, "mockResponse", resp)
}

// Test LoggingServerInterceptor with Error.
func TestLoggingServerInterceptorWithError(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()

	interceptor := grpcserver.LoggingServerInterceptor(logger)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}

	mockErrorHandler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return nil, status.Error(codes.Internal, "internal error")
	}

	resp, err := interceptor(context.Background(), "testRequest", info, mockErrorHandler)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}
