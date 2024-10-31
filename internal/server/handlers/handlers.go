package handlers

import (
	"time"

	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/validators"
)

// Define a constant for the timeout duration.
const (
	defaultTimeout = 500 * time.Millisecond // Default timeout for metrics handler
)

type MetricHandler struct {
	validator validators.MValidator
	logger    *zerolog.Logger
	repo      model.Repository
	timeout   time.Duration
}

// NewMetricsHandler - constructor for MetricsHandler.
func NewMetricsHandler(repo model.Repository, l *zerolog.Logger) *MetricHandler {
	return &MetricHandler{
		validator: validators.NewMetricsValidator(),
		logger:    l,
		repo:      repo,
		timeout:   defaultTimeout,
	}
}
