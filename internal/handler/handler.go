package handler

import (
	. "github.com/npavlov/go-metrics-service/internal/metric-types"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"net/http"
	"strconv"
)

func GetUpdateHandler(ms storage.Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Пример: /update/gauge/metric_name/123.4
		parts := splitURL(r.URL.Path)
		if len(parts) != 4 {
			http.Error(w, "invalid URL format", http.StatusNotFound)
			return
		}

		metricType := MetricType(parts[1])
		metricName := MetricName(parts[2])
		metricValue := parts[3]

		// metric name not found
		if len(metricName) < 1 {
			http.Error(w, "no metric id", http.StatusBadRequest)
			return
		}
		// metric value not found TODO: improve
		if len(metricValue) == 0 {
			http.Error(w, "no value", http.StatusBadRequest)
			return
		}

		switch metricType {
		case Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "invalid gauge value", http.StatusBadRequest)
				return
			}
			ms.UpdateGauge(metricName, value)
		case Counter:
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

// Разделение пути URL
func splitURL(url string) []string {
	return filterEmptyStrings(splitString(url, '/'))
}

// Вспомогательная функция для разделения строки
func splitString(s string, sep rune) []string {
	var result []string
	current := ""
	for _, char := range s {
		if char == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	result = append(result, current)
	return result
}

// Фильтрация пустых строк
func filterEmptyStrings(s []string) []string {
	var result []string
	for _, v := range s {
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}
