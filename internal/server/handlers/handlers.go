package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/npavlov/go-metrics-service/internal/server/templates"
	"github.com/rs/zerolog/log"
)

type Handlers interface {
	Render(w http.ResponseWriter, r *http.Request)
	Retrieve(w http.ResponseWriter, r *http.Request)
	RetrieveModel(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	UpdateModel(w http.ResponseWriter, r *http.Request)
	SetRouter()
}

type MetricHandler struct {
	router *chi.Mux
	st     storage.Repository
}

// NewMetricsHandler - constructor for MetricsHandler.
func NewMetricsHandler(st storage.Repository, router *chi.Mux) *MetricHandler {
	return &MetricHandler{router: router, st: st}
}

func (mh *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
	metricType := domain.MetricType(chi.URLParam(r, "metricType"))
	metricName := domain.MetricName(chi.URLParam(r, "metricName"))
	metricValue := chi.URLParam(r, "value")

	metric := &model.Metric{
		ID: metricName,
	}
	switch metricType {
	case domain.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			log.Error().Err(err).Msg("error parsing float value")
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}
		metric.MType = metricType
		metric.Value = &value
	case domain.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			log.Error().Err(err).Msg("error parsing int64 value")
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}
		metric.MType = metricType
		metric.Delta = &value
	}

	if err := mh.st.Update(metric); err != nil {
		log.Error().Err(err).Msg("error updating metric")
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	l := logger.NewLogger().Get()

	// Decode the incoming JSON request into the Metric struct
	var metric *model.Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		l.Error().Err(err).Msg("UpdateModel: Invalid JSON input")
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)

		return
	}

	// Update the metric in the storage and retrieve the updated values
	err := mh.st.Update(metric)
	if err != nil {
		l.Error().Err(err).Msgf("UpdateModel: Failed to update metric %s", metric.ID)
		http.Error(w, "Failed to update metric", http.StatusBadRequest)

		return
	}

	// Prepare the updated metric to be returned
	responseMetric, found := mh.st.Get(metric.ID)
	if found != true {
		l.Error().Msgf("UpdateModel: Failed to retrieve model from memory %s", metric.ID)
		http.Error(w, "Failed to retrieve model from memory", http.StatusBadRequest)

		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(responseMetric)
	if err != nil {
		l.Error().Err(err).Msg("Failed to encode response JSON")
		http.Error(w, "Failed to process response", http.StatusInternalServerError)
	}
}

func (mh *MetricHandler) Retrieve(w http.ResponseWriter, r *http.Request) {

	metricName := domain.MetricName(chi.URLParam(r, "metricName"))

	metricModel, found := mh.st.Get(metricName)
	if found != true {
		log.Error().Msgf("Retrieve: Failed to retrieve model from memory %s", metricName)
		http.Error(w, "Failed to retrieve model from memory", http.StatusNotFound)

		return
	}

	_, _ = w.Write([]byte(metricModel.GetValue()))
	w.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) RetrieveModel(w http.ResponseWriter, r *http.Request) {
	l := logger.NewLogger().Get()

	// Decode the incoming JSON request into the Metric struct
	var metric *model.Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		l.Error().Err(err).Msg("UpdateModel: Invalid JSON input")
		http.Error(w, "Invalid JSON input", http.StatusNotFound)

		return
	}

	// Prepare the updated metric to be returned
	responseMetric, found := mh.st.Get(metric.ID)
	if found != true {
		l.Error().Msgf("UpdateModel: Failed to retrieve model from memory %s", metric.ID)
		http.Error(w, "Failed to retrieve model from memory", http.StatusNotFound)

		return
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(responseMetric)
	if err != nil {
		l.Error().Err(err).Msg("Failed to encode response JSON")
		http.Error(w, "Failed to process response", http.StatusInternalServerError)
	}
}

func (mh *MetricHandler) Render(w http.ResponseWriter, _ *http.Request) {
	l := logger.NewLogger().Get()

	page := struct {
		Metrics map[domain.MetricName]model.Metric
	}{
		Metrics: mh.st.GetAll(),
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

// SetRouter Embedding middleware setup in the constructor.
func (mh *MetricHandler) SetRouter() {
	mh.router.Use(middlewares.LoggingMiddleware)
	mh.router.Use(middleware.Recoverer)
	mh.router.Use(middlewares.GzipMiddleware)
	mh.router.Use(middlewares.BrotliMiddleware)
	mh.router.Use(middlewares.GzipDecompressionMiddleware)

	mh.router.Route("/", func(r chi.Router) {
		r.Route("/", func(r chi.Router) {
			r.With(middlewares.ContentMiddleware("text/html")).Get("/", mh.Render)
		})
		r.Route("/update", func(r chi.Router) {
			r.With(middlewares.ContentMiddleware("application/json")).Post("/", mh.UpdateModel)
		})
		r.Route("/update/{metricType}/{metricName}/{value}", func(r chi.Router) {
			r.With(middlewares.ContentMiddleware("application/text")).Post("/", mh.Update)
		})
		r.Route("/value", func(r chi.Router) {
			r.With(middlewares.ContentMiddleware("application/json")).Post("/", mh.RetrieveModel)
		})
		r.Route("/value/{metricType}/{metricName}", func(r chi.Router) {
			r.With(middlewares.ContentMiddleware("application/text")).Get("/", mh.Retrieve)
		})
	})
}
