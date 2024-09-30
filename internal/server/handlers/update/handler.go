package update

import (
	"github.com/go-chi/chi/v5"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
	"net/http"
)

func GetUpdateHandler(ms storage.Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := types.MetricType(chi.URLParam(r, "metricType"))
		metricName := types.MetricName(chi.URLParam(r, "metricName"))
		metricValue := chi.URLParam(r, "value")

		if err := ms.UpdateMetric(metricType, metricName, metricValue); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
