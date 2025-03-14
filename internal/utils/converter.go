package utils

import (
	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

func FromGModelToDBModel(metric *pb.Metric) *db.Metric {
	var newMetric *db.Metric
	switch metric.GetMtype() {
	case pb.Metric_TYPE_COUNTER:
		delta := metric.GetDelta()
		newMetric = db.NewMetric(domain.MetricName(metric.GetId()), domain.Counter, &delta, nil)
	case pb.Metric_TYPE_GAUGE:
		val := metric.GetValue()
		newMetric = db.NewMetric(domain.MetricName(metric.GetId()), domain.Gauge, nil, &val)
	case pb.Metric_TYPE_UNSPECIFIED:
		return nil
	}

	return newMetric
}

func FromDBModelToGModel(metric *db.Metric) *pb.Metric {
	//nolint:exhaustruct
	newMetric := &pb.Metric{
		Id: string(metric.ID),
	}

	switch metric.MType {
	case domain.Counter:
		newMetric.Delta = metric.Delta
		newMetric.Mtype = pb.Metric_TYPE_COUNTER
	case domain.Gauge:
		newMetric.Value = metric.Value
		newMetric.Mtype = pb.Metric_TYPE_GAUGE
	}

	return newMetric
}
