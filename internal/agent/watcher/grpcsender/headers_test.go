package grpcsender_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher/grpcsender"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

func TestHeadersInterceptor(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()
	cfg := &config.Config{Key: "test-key"}
	ip := "127.0.0.1"
	interceptor := grpcsender.HeadersInterceptor(cfg, ip, logger)

	t.Run("adds X-Real-IP header", func(t *testing.T) {
		t.Parallel()

		invoker := &MockUnaryInvoker{}
		invoker.On("Invoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Run(func(args mock.Arguments) {
			ctx, ok := args.Get(0).(context.Context)
			require.True(t, ok)
			md, ok := metadata.FromOutgoingContext(ctx)
			require.True(t, ok)
			assert.Contains(t, md.Get("X-Real-IP"), ip)
		})

		err := interceptor(context.Background(), "testMethod", "testRequest", "testReply", nil, invoker.Invoke)
		assert.NoError(t, err)
	})

	t.Run("adds HashSHA256 header when key is set", func(t *testing.T) {
		t.Parallel()

		req := &pb.SetMetricRequest{Metric: &pb.Metric{Id: "test", Mtype: pb.Metric_TYPE_GAUGE, Value: float64Ptr(1)}}
		payload, err := utils.MarshalProtoMessage(req)
		require.NoError(t, err)
		hash := utils.CalculateHash(cfg.Key, payload)

		invoker := &MockUnaryInvoker{}
		invoker.On("Invoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Run(func(args mock.Arguments) {
			ctx, ok := args.Get(0).(context.Context)
			require.True(t, ok)
			md, ok := metadata.FromOutgoingContext(ctx)
			require.True(t, ok)
			hashes := md.Get("HashSHA256")
			assert.Contains(t, hashes[0], hash)
		})

		err = interceptor(context.Background(), "testMethod", req, "testReply", nil, invoker.Invoke)
		assert.NoError(t, err)
	})

	t.Run("handles serialization failure gracefully", func(t *testing.T) {
		t.Parallel()

		cfgWithInvalidKey := &config.Config{Key: ""} // Key is empty, should not generate hash
		interceptor := grpcsender.HeadersInterceptor(cfgWithInvalidKey, ip, logger)

		//nolint:err113
		serializationErr := errors.New("serialization failed")

		invoker := &MockUnaryInvoker{}
		invoker.On("Invoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(serializationErr)

		err := interceptor(context.Background(), "testMethod", struct{}{}, "testReply", nil, invoker.Invoke)
		require.Error(t, err)
		assert.Equal(t, "serialization failed", err.Error())
	})
}
