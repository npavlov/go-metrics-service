package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/npavlov/go-metrics-service/internal/server/templates"
	"github.com/rs/zerolog/log"
	"net/http"
)

type Handlers interface {
	Render(http.ResponseWriter, *http.Request)
	Retrieve(http.ResponseWriter, *http.Request)
	RetrieveModel(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	UpdateModel(http.ResponseWriter, *http.Request)
	JSONResponse(w http.ResponseWriter, status int, v interface{})
	SetRouter()
}

type MetricHandler struct {
	router *chi.Mux
	st     storage.Repository
}

// NewMetricsHandler - constructor for MetricsHandler
func NewMetricsHandler(st storage.Repository, router *chi.Mux) *MetricHandler {
	mh := MetricHandler{
		router: router,
		st:     st,
	}
	mh.setRouter()
	return &mh
}

func (mh *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
	metricType := domain.MetricType(chi.URLParam(r, "metricType"))
	metricName := domain.MetricName(chi.URLParam(r, "metricName"))
	metricValue := chi.URLParam(r, "value")

	if err := mh.st.UpdateMetric(metricType, metricName, metricValue); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	l := logger.Get()

	// Decode the incoming JSON request into the Metric struct
	var metric *model.Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		l.Error().Err(err).Msg("UpdateModel: Invalid JSON input")
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Update the metric in the storage and retrieve the updated values
	err := mh.st.UpdateMetricModel(metric)
	if err != nil {
		l.Error().Err(err).Msgf("UpdateModel: Failed to update metric %s", metric.ID)
		http.Error(w, "Failed to update metric", http.StatusBadRequest)
		return
	}

	// Prepare the updated metric to be returned
	responseMetric, err := mh.st.GetMetricModel(metric)
	if err != nil {
		l.Error().Err(err).Msgf("UpdateModel: Failed to retrieve model from memory %s", metric.ID)
		http.Error(w, "Failed to retrieve model from memory", http.StatusBadRequest)
		return
	}

	mh.JSONResponse(w, http.StatusOK, responseMetric)
}

func (mh *MetricHandler) Retrieve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/text")
	metricType := domain.MetricType(chi.URLParam(r, "metricType"))
	metricName := domain.MetricName(chi.URLParam(r, "metricName"))

	metric := &model.Metric{
		MType: metricType,
		ID:    metricName,
	}

	metricModel, err := mh.st.GetMetricModel(metric)
	if err != nil {
		log.Error().Err(err).Msgf("Retrieve: Failed to retrieve model from memory %s", metric.ID)
		http.Error(w, "Failed to retrieve model from memory", http.StatusNotFound)
		return
	}

	_, _ = w.Write([]byte(metricModel.GetValue()))
	w.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) RetrieveModel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	l := logger.Get()

	// Decode the incoming JSON request into the Metric struct
	var metric *model.Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		l.Error().Err(err).Msg("UpdateModel: Invalid JSON input")
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Prepare the updated metric to be returned
	responseMetric, err := mh.st.GetMetricModel(metric)
	if err != nil {
		l.Error().Err(err).Msgf("UpdateModel: Failed to retrieve model from memory %s", metric.ID)
		http.Error(w, "Failed to retrieve model from memory", http.StatusBadRequest)
		return
	}

	mh.JSONResponse(w, http.StatusOK, responseMetric)
}

func (mh *MetricHandler) Render(w http.ResponseWriter, _ *http.Request) {
	l := logger.Get()

	page := struct {
		Gauges   map[domain.MetricName]float64
		Counters map[domain.MetricName]int64
	}{
		Gauges:   mh.st.GetGauges(),
		Counters: mh.st.GetCounters(),
	}

	reader := templates.NewEmbedReader()
	tmpl, err := reader.Read("index.html")
	if err != nil {
		l.Error().Err(err).Msg("Could not load template")
		http.Error(w, "Failed to load template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, page); err != nil {
		l.Error().Err(err).Msg("Could not render template")
		http.Error(w, "Failed to render page: "+err.Error(), http.StatusInternalServerError)
	}
}

// JSONResponse writes a JSON response with appropriate headers and status code
func (mh *MetricHandler) JSONResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Get().Error().Err(err).Msg("Failed to encode response JSON")
		http.Error(w, "Failed to process response", http.StatusInternalServerError)
	}
}

// Embedding middleware setup in the constructor
func (mh *MetricHandler) setRouter() {
	mh.router.Use(middlewares.LoggingMiddleware)
	mh.router.Use(middleware.Recoverer)

	mh.router.Route("/", func(r chi.Router) {
		r.Get("/", mh.Render)
		r.Route("/update", func(r chi.Router) {
			r.Post("/", mh.UpdateModel)
			r.Post("/{metricType}/{metricName}/{value}", mh.Update)
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/", mh.RetrieveModel)
			r.Get("/{metricType}/{metricName}", mh.Retrieve)
		})
	})
}
