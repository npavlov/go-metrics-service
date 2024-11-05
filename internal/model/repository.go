package model

import (
	"context"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

type Repository interface {
	Get(context context.Context, name domain.MetricName) (*db.MtrMetric, bool)
	GetMany(context context.Context, names []domain.MetricName) (map[domain.MetricName]db.MtrMetric, error)
	Create(context context.Context, metric *db.MtrMetric) error
	GetAll(context context.Context) map[domain.MetricName]db.MtrMetric
	Update(context context.Context, metric *db.MtrMetric) error
	UpdateMany(context context.Context, metrics *[]db.MtrMetric) error
}
