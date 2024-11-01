package model

import (
	"context"

	"github.com/npavlov/go-metrics-service/internal/domain"
)

type Repository interface {
	Get(context context.Context, name domain.MetricName) (*Metric, bool)
	GetMany(context context.Context, names []domain.MetricName) (map[domain.MetricName]Metric, error)
	Create(context context.Context, metric *Metric) error
	GetAll(context context.Context) map[domain.MetricName]Metric
	Update(context context.Context, metric *Metric) error
	UpdateMany(context context.Context, metrics *[]Metric) error
}
