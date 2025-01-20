package handlers

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/validators"
	"github.com/npavlov/go-metrics-service/web"
)

type MetricHandler struct {
	validator   validators.MValidator
	logger      *zerolog.Logger
	repo        model.Repository
	embedReader *web.EmbedReader
	json        jsoniter.API
}

// NewMetricsHandler - constructor for MetricsHandler.
func NewMetricsHandler(repo model.Repository, l *zerolog.Logger) *MetricHandler {
	return &MetricHandler{
		validator:   validators.NewMetricsValidator(),
		logger:      l,
		repo:        repo,
		embedReader: web.NewEmbedReader(),
		json:        jsoniter.ConfigCompatibleWithStandardLibrary,
	}
}
