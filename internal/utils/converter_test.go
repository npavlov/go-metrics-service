package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

func TestFromGModelToDBModel(t *testing.T) {
	t.Parallel()

	metricCounter := &pb.Metric{
		Id:    "counter_metric",
		Mtype: pb.Metric_TYPE_COUNTER,
		Delta: int64Ptr(42),
	}

	dbMetric := utils.FromGModelToDBModel(metricCounter)
	assert.NotNil(t, dbMetric)
	assert.Equal(t, domain.Counter, dbMetric.MType)
	assert.Equal(t, int64(42), *dbMetric.Delta)

	metricGauge := &pb.Metric{
		Id:    "gauge_metric",
		Mtype: pb.Metric_TYPE_GAUGE,
		Value: float64Ptr(3.14),
	}

	dbMetric = utils.FromGModelToDBModel(metricGauge)
	assert.NotNil(t, dbMetric)
	assert.Equal(t, domain.Gauge, dbMetric.MType)
	assert.InDelta(t, 3.14, *dbMetric.Value, 0.001)
}

func TestFromDBModelToGModel(t *testing.T) {
	t.Parallel()

	dbMetricCounter := db.NewMetric("counter_metric", domain.Counter, int64Ptr(42), nil)

	gMetric := utils.FromDBModelToGModel(dbMetricCounter)
	assert.NotNil(t, gMetric)
	assert.Equal(t, pb.Metric_TYPE_COUNTER, gMetric.GetMtype())
	assert.Equal(t, int64(42), gMetric.GetDelta())

	dbMetricGauge := db.NewMetric("counter_metric", domain.Gauge, nil, float64Ptr(3.14))

	gMetric = utils.FromDBModelToGModel(dbMetricGauge)
	assert.NotNil(t, gMetric)
	assert.Equal(t, pb.Metric_TYPE_GAUGE, gMetric.GetMtype())
	assert.InDelta(t, 3.14, gMetric.GetValue(), 0.001)
}

// Helper function to create float64 pointer.
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create int64 pointer.
func int64Ptr(i int64) *int64 {
	return &i
}
