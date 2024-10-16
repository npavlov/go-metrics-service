package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
	"github.com/npavlov/go-metrics-service/internal/server/templates"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"net/http"
	"strconv"
)

type Handlers interface {
	Render(http.ResponseWriter, *http.Request)
	Retrieve(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
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

func (mh *MetricHandler) Retrieve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/text")
	metricType := domain.MetricType(chi.URLParam(r, "metricType"))
	metricName := domain.MetricName(chi.URLParam(r, "metricName"))

	switch metricType {
	case domain.Gauge:
		if value, found := mh.st.GetGauge(metricName); found {
			_, _ = w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
			return
		}
		http.Error(w, "unknown metric name", http.StatusNotFound)
	case domain.Counter:
		if value, found := mh.st.GetCounter(metricName); found {
			_, _ = w.Write([]byte(strconv.FormatInt(value, 10)))
			return
		}
		http.Error(w, "unknown metric name", http.StatusNotFound)
	default:
		http.Error(w, "unknown metric type", http.StatusNotFound)
	}
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

// Embedding middleware setup in the constructor
func (mh *MetricHandler) setRouter() {
	mh.router.Use(middlewares.LoggingMiddleware)
	mh.router.Use(middleware.Recoverer)

	mh.router.Route("/", func(r chi.Router) {
		r.Get("/", mh.Render)
		r.Route("/update", func(r chi.Router) {
			r.Post("/{metricType}/{metricName}/{value}", mh.Update)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{metricType}/{metricName}", mh.Retrieve)
		})
	})
}
