package handlers

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/npavlov/go-metrics-service/internal/validators"
)

type Handlers interface {
	Render(w http.ResponseWriter, r *http.Request)
	Retrieve(w http.ResponseWriter, r *http.Request)
	RetrieveModel(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	UpdateModel(w http.ResponseWriter, r *http.Request)
}

type MetricHandler struct {
	st        storage.Repository
	validator validators.MValidator
	logger    *zerolog.Logger
}

// NewMetricsHandler - constructor for MetricsHandler.
func NewMetricsHandler(st storage.Repository, l *zerolog.Logger) *MetricHandler {
	return &MetricHandler{
		st:        st,
		validator: validators.NewMetricsValidator(),
		logger:    l,
	}
}
