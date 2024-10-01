package handlers

import (
	"embed"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
	"html/template"
	"log"
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

//go:embed templates/*
var tplFolder embed.FS

// NewMetricsHandler - constructor for MetricsHandler
func NewMetricsHandler(st storage.Repository, router *chi.Mux) *MetricHandler {
	return &MetricHandler{
		router: router,
		st:     st,
	}
}

func (mh *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
	metricType := types.MetricType(chi.URLParam(r, "metricType"))
	metricName := types.MetricName(chi.URLParam(r, "metricName"))
	metricValue := chi.URLParam(r, "value")

	fmt.Println(metricValue, metricName, metricValue)

	if err := mh.st.UpdateMetric(metricType, metricName, metricValue); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) Retrieve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/text")
	metricType := types.MetricType(chi.URLParam(r, "metricType"))
	metricName := types.MetricName(chi.URLParam(r, "metricName"))

	switch metricType {
	case types.Gauge:
		if value, found := mh.st.GetGauge(metricName); found {
			_, _ = w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
			return
		}
		http.Error(w, "unknown metric name", http.StatusNotFound)
	case types.Counter:
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
	type MetricsPage struct {
		Gauges   map[types.MetricName]float64
		Counters map[types.MetricName]int64
	}

	page := MetricsPage{
		Gauges:   mh.st.GetGauges(),
		Counters: mh.st.GetCounters(),
	}

	// Parse the template files from the embedded filesystem
	tmpl, err := template.ParseFS(tplFolder, "templates/*.html")
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (mh *MetricHandler) SetRouter() {
	// Useful middlewares, extra logging
	mh.router.Use(middleware.Logger)
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
