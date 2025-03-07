package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"github.com/npavlov/go-metrics-service/internal/server/grpc"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestSubnetInterceptor(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()
	subnet := "192.168.1.0/24"

	interceptor := grpc.SubnetInterceptor(subnet, logger)

	md := metadata.Pairs("X-Real-IP", "192.168.1.10")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor(ctx, "testRequest", nil, mockHandler)

	require.NoError(t, err)
	assert.Equal(t, "mockResponse", resp)
}

// Unauthorized access test.
func TestSubnetInterceptorUnauthorized(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()
	subnet := "192.168.1.0/24"

	interceptor := grpc.SubnetInterceptor(subnet, logger)

	md := metadata.Pairs("X-Real-IP", "10.10.10.10")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor(ctx, "testRequest", nil, mockHandler)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "unauthorized access")
}
