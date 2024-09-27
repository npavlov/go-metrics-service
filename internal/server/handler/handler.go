package handler

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	types "github.com/npavlov/go-metrics-service/internal/agent/metrictypes"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"net/http"
	"strconv"
)

func GetUpdateHandler(ms storage.Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := types.MetricType(chi.URLParam(r, "metricType"))
		metricName := types.MetricName(chi.URLParam(r, "metricName"))
		metricValue := chi.URLParam(r, "value")

		fmt.Println(metricValue, metricType, metricName)

		switch metricType {
		case types.Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "invalid gauge value", http.StatusBadRequest)
				return
			}
			ms.UpdateGauge(metricName, value)
		case types.Counter:
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "invalid counter value", http.StatusBadRequest)
				return
			}
			ms.UpdateCounter(metricName, value)
		default:
			http.Error(w, "unknown metric type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
