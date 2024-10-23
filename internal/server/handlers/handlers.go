package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/npavlov/go-metrics-service/internal/server/templates"
	validators "github.com/npavlov/go-metrics-service/internal/validators"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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
	router    *chi.Mux
	st        storage.Repository
	validator validators.MValidator
	l         *zerolog.Logger
}

// NewMetricsHandler - constructor for MetricsHandler.
func NewMetricsHandler(st storage.Repository, router *chi.Mux, l *zerolog.Logger) *MetricHandler {
	return &MetricHandler{
		router:    router,
		st:        st,
		validator: validators.NewMetricsValidator(),
		l:         l,
	}
}

func (mh *MetricHandler) Update(response http.ResponseWriter, request *http.Request) {
	metricType := domain.MetricType(chi.URLParam(request, "metricType"))
	metricName := domain.MetricName(chi.URLParam(request, "metricName"))
	metricValue := chi.URLParam(request, "value")

	newMetric, err := mh.validator.FromVars(metricName, metricType, metricValue)
	if err != nil {
		mh.l.Error().Err(err).Msg("error validating structure")
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	_, err = mh.updateAndReturn(newMetric)
	if err != nil {
		mh.l.Error().Err(err).Msg("error updating metric")
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	response.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) UpdateModel(response http.ResponseWriter, request *http.Request) {
	newMetric, err := mh.validator.FromBody(request.Body)
	if err != nil {
		mh.l.Error().Err(err).Msg("error validating structure")
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	metric, err := mh.updateAndReturn(newMetric)
	if err != nil {
		mh.l.Error().Err(err).Msg("error updating metric")
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	response.WriteHeader(http.StatusOK)
	err = json.NewEncoder(response).Encode(metric)
	if err != nil {
		mh.l.Error().Err(err).Msg("Failed to encode response JSON")
		http.Error(response, "Failed to process response", http.StatusInternalServerError)
	}
}

func (mh *MetricHandler) updateAndReturn(newMetric *model.Metric) (*model.Metric, error) {
	existingMetric, found := mh.st.Get(newMetric.ID)

	if found {
		existingMetric.SetValue(newMetric.Delta, newMetric.Value)

		err := mh.st.Update(existingMetric)
		if err != nil {
			mh.l.Error().Err(err).Msg("error updating existingMetric")

			return nil, errors.Wrap(err, "error updating existingMetric")
		}

		return existingMetric, nil
	}

	err := mh.st.Create(newMetric)
	if err != nil {
		mh.l.Error().Err(err).Msg("error creating Metric")

		return nil, errors.Wrap(err, "error creating new Metric")
	}

	return newMetric, nil
}

func (mh *MetricHandler) Retrieve(response http.ResponseWriter, request *http.Request) {
	metricName := domain.MetricName(chi.URLParam(request, "metricName"))

	metricModel, found := mh.st.Get(metricName)
	if !found {
		log.Error().Msgf("Retrieve: Failed to retrieve model from memory %s", metricName)
		http.Error(response, "Failed to retrieve model from memory", http.StatusNotFound)

		return
	}

	_, _ = response.Write([]byte(metricModel.GetValue()))
	response.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) RetrieveModel(response http.ResponseWriter, request *http.Request) {
	// Decode the incoming JSON request into the Metric struct
	var metric *model.Metric
	if err := json.NewDecoder(request.Body).Decode(&metric); err != nil {
		mh.l.Error().Err(err).Msg("UpdateModel: Invalid JSON input")
		http.Error(response, "Invalid JSON input", http.StatusNotFound)

		return
	}

	// Prepare the updated metric to be returned
	responseMetric, found := mh.st.Get(metric.ID)
	if !found {
		mh.l.Error().Msgf("UpdateModel: Failed to retrieve model from memory %s", metric.ID)
		http.Error(response, "Failed to retrieve model from memory", http.StatusNotFound)

		return
	}

	response.WriteHeader(http.StatusOK)
	err := json.NewEncoder(response).Encode(responseMetric)
	if err != nil {
		mh.l.Error().Err(err).Msg("Failed to encode response JSON")
		http.Error(response, "Failed to process response", http.StatusInternalServerError)
	}
}

func (mh *MetricHandler) Render(response http.ResponseWriter, _ *http.Request) {
	page := struct {
		Metrics map[domain.MetricName]model.Metric
	}{
		Metrics: mh.st.GetAll(),
	}

	reader := templates.NewEmbedReader()
	tmpl, err := reader.Read("index.html")
	if err != nil {
		mh.l.Error().Err(err).Msg("Could not load template")
		http.Error(response, "Failed to load template: "+err.Error(), http.StatusInternalServerError)

		return
	}

	if err := tmpl.Execute(response, page); err != nil {
		mh.l.Error().Err(err).Msg("Could not render template")
		http.Error(response, "Failed to render page: "+err.Error(), http.StatusInternalServerError)
	}
}

// SetRouter Embedding middleware setup in the constructor.
func (mh *MetricHandler) SetRouter() {
	mh.router.Use(middlewares.LoggingMiddleware(mh.l))
	mh.router.Use(middleware.Recoverer)
	mh.router.Use(middlewares.GzipMiddleware)
	mh.router.Use(middlewares.BrotliMiddleware)
	mh.router.Use(middlewares.GzipDecompressionMiddleware)

	mh.router.Route("/", func(router chi.Router) {
		router.Route("/", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("text/html")).
				Get("/", mh.Render)
		})
		router.Route("/update", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("application/json")).
				Post("/", mh.UpdateModel)
		})
		router.Route("/update/{metricType}/{metricName}/{value}", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("application/text")).
				Post("/", mh.Update)
		})
		router.Route("/value", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("application/json")).
				Post("/", mh.RetrieveModel)
		})
		router.Route("/value/{metricType}/{metricName}", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("application/text")).
				Get("/", mh.Retrieve)
		})
	})
}
