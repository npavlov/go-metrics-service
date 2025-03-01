package utils

import (
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	pb "github.com/npavlov/go-metrics-service/proto/v1"
)

func FromGModelToDBModel(metric *pb.Metric) *db.Metric {
	var newMetric *db.Metric
	switch metric.GetMtype() {
	case pb.Metric_COUNTER:
		newMetric = db.NewMetric(domain.MetricName(metric.GetId()), domain.Counter, &metric.Delta, nil)
	case pb.Metric_GAUGE:
		newMetric = db.NewMetric(domain.MetricName(metric.GetId()), domain.Gauge, nil, &metric.Value)
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
		newMetric.Delta = *metric.Delta
		newMetric.Mtype = pb.Metric_COUNTER
	case domain.Gauge:
		newMetric.Value = *metric.Value
		newMetric.Mtype = pb.Metric_GAUGE
	}

	return newMetric
}
