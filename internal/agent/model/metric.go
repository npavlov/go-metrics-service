package model

import (
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

type Metric struct {
	db.Metric
	MSource domain.MetricSource
	MAlias  domain.MetricAlias
}
