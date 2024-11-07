package model

import (
	"context"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

type Repository interface {
	Get(context context.Context, name domain.MetricName) (*db.Metric, bool)
	GetMany(context context.Context, names []domain.MetricName) (map[domain.MetricName]db.Metric, error)
	Create(context context.Context, metric *db.Metric) error
	GetAll(context context.Context) map[domain.MetricName]db.Metric
	Update(context context.Context, metric *db.Metric) error
	UpdateMany(context context.Context, metrics *[]db.Metric) error
}
