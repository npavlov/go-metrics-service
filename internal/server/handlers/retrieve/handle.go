package retrieve

import (
	"github.com/go-chi/chi/v5"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
	"net/http"
	"strconv"
)

// Helper function for sending error responses
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}

func GetRetrieveHandler(ms storage.Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/text")
		metricType := types.MetricType(chi.URLParam(r, "metricType"))
		metricName := types.MetricName(chi.URLParam(r, "metricName"))

		switch metricType {
		case types.Gauge:
			if value, found := ms.GetGauge(metricName); found {
				w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
				return
			}
			sendErrorResponse(w, "unknown metric name", http.StatusNotFound)
		case types.Counter:
			if value, found := ms.GetCounter(metricName); found {
				w.Write([]byte(strconv.FormatInt(value, 10)))
				return
			}
			sendErrorResponse(w, "unknown metric name", http.StatusNotFound)
		default:
			sendErrorResponse(w, "unknown metric type", http.StatusNotFound)
		}
	}
}
