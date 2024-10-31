package handlers

import (
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/validators"
)

type MetricHandler struct {
	validator validators.MValidator
	logger    *zerolog.Logger
	sMonitor  *storage.StorageMonitor
}

// NewMetricsHandler - constructor for MetricsHandler.
func NewMetricsHandler(sMonitor *storage.StorageMonitor, l *zerolog.Logger) *MetricHandler {
	return &MetricHandler{
		validator: validators.NewMetricsValidator(),
		logger:    l,
		sMonitor:  sMonitor,
	}
}
