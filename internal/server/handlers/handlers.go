package handlers

import (
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/server/repository"
	"github.com/npavlov/go-metrics-service/internal/validators"
)

type MetricHandler struct {
	validator   validators.MValidator
	logger      *zerolog.Logger
	universalDB repository.Universal
}

// NewMetricsHandler - constructor for MetricsHandler.
func NewMetricsHandler(universalDB repository.Universal, l *zerolog.Logger) *MetricHandler {
	return &MetricHandler{
		validator:   validators.NewMetricsValidator(),
		logger:      l,
		universalDB: universalDB,
	}
}
