package handlers

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/validators"
	"github.com/npavlov/go-metrics-service/web"
)

// MetricHandler handles requests related to metrics.
type MetricHandler struct {
	validator   validators.MValidator // Validator for metric inputs.
	logger      *zerolog.Logger       // Logger for logging errors and info.
	repo        model.Repository      // Repository for accessing metric data.
	embedReader *web.EmbedReader      // Reader for embedded templates.
	json        jsoniter.API          // JSON API for encoding/decoding JSON data.
}

// NewMetricsHandler creates and initializes a new instance of MetricHandler.
//
// Parameters:
//   - repo: The repository to manage metric data.
//   - l: The logger for logging messages.
//
// Returns:
//   - A pointer to a new MetricHandler instance.
func NewMetricsHandler(repo model.Repository, l *zerolog.Logger) *MetricHandler {
	return &MetricHandler{
		validator:   validators.NewMetricsValidator(),
		logger:      l,
		repo:        repo,
		embedReader: web.NewEmbedReader(),
		json:        jsoniter.ConfigCompatibleWithStandardLibrary,
	}
}
