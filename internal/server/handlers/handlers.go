package handlers

import (
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/validators"
)

type MetricHandler struct {
	validator validators.MValidator
	logger    *zerolog.Logger
	repo      model.Repository
}

// NewMetricsHandler - constructor for MetricsHandler.
func NewMetricsHandler(repo model.Repository, l *zerolog.Logger) *MetricHandler {
	return &MetricHandler{
		validator: validators.NewMetricsValidator(),
		logger:    l,
		repo:      repo,
	}
}
