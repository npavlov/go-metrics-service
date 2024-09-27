package update

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
	"net/http"
	"strconv"
)

const (
	errInvalidGauge   = "invalid gauge value"
	errInvalidCounter = "invalid counter value"
	errUnknownMetric  = "unknown metric type"
)

func GetUpdateHandler(ms storage.Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := types.MetricType(chi.URLParam(r, "metricType"))
		metricName := types.MetricName(chi.URLParam(r, "metricName"))
		metricValue := chi.URLParam(r, "value")

		if err := updateMetric(ms, metricType, metricName, metricValue); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func updateMetric(ms storage.Repository, metricType types.MetricType, metricName types.MetricName, metricValue string) error {
	switch metricType {
	case types.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return errors.New(errInvalidGauge)
		}
		ms.UpdateGauge(metricName, value)
	case types.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return errors.New(errInvalidCounter)
		}
		ms.UpdateCounter(metricName, value)
	default:
		return errors.New(errUnknownMetric)
	}
	return nil
}
