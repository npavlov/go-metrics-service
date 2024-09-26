package handler

import (
	"github.com/npavlov/go-metrics-service/internal/storage"
	"net/http"
	"strconv"
)

func GetUpdateHandler(ms *storage.MemStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Пример: /update/gauge/metric_name/123.4
		parts := splitURL(r.URL.Path)
		if len(parts) != 4 {
			http.Error(w, "invalid URL format", http.StatusBadRequest)
			return
		}

		metricType := storage.MetricType(parts[1])
		metricName := parts[2]
		metricValue := parts[3]

		if len(metricName) < 2 {
			http.Error(w, "unknown metric id", http.StatusBadRequest)
			return
		}

		switch metricType {
		case storage.Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "invalid gauge value", http.StatusBadRequest)
				return
			}
			ms.UpdateGauge(metricName, value)
		case storage.Counter:
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
