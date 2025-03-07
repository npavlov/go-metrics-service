package model

import (
	"context"

	"github.com/npavlov/go-metrics-service/internal/server/db"
)

type Sender interface {
	SendMetricsBatch(ctx context.Context, metrics []db.Metric) ([]db.Metric, error)
	SendMetric(ctx context.Context, metric db.Metric) (*db.Metric, error)
	Close()
}
