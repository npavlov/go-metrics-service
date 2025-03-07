package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

func TestMarshalProtoMessage(t *testing.T) {
	t.Parallel()

	metric := &pb.Metric{
		Id:    "test_metric",
		Mtype: pb.Metric_TYPE_COUNTER,
		Delta: int64Ptr(42),
	}

	data, err := utils.MarshalProtoMessage(metric)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestUnmarshalProtoMessage(t *testing.T) {
	t.Parallel()

	metric := &pb.Metric{
		Id:    "test_metric",
		Mtype: pb.Metric_TYPE_COUNTER,
		Delta: int64Ptr(42),
	}

	data, _ := utils.MarshalProtoMessage(metric)

	var newMetric pb.Metric
	err := utils.UnmarshalProtoMessage(data, &newMetric)
	require.NoError(t, err)
	assert.Equal(t, metric.GetId(), newMetric.GetId())
	assert.Equal(t, metric.GetMtype(), newMetric.GetMtype())
	assert.Equal(t, metric.GetDelta(), newMetric.GetDelta())
}
