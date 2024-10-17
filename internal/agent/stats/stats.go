package stats

import (
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"reflect"
)

type Stats struct {
	Alloc         float64 `metricType:"gauge" metricSource:"runtime"`
	BuckHashSys   float64 `metricType:"gauge" metricSource:"runtime"`
	Frees         float64 `metricType:"gauge" metricSource:"runtime"`
	GCCPUFraction float64 `metricType:"gauge" metricSource:"runtime"`
	GCSys         float64 `metricType:"gauge" metricSource:"runtime"`
	HeapAlloc     float64 `metricType:"gauge" metricSource:"runtime"`
	HeapIdle      float64 `metricType:"gauge" metricSource:"runtime"`
	HeapInuse     float64 `metricType:"gauge" metricSource:"runtime"`
	HeapObjects   float64 `metricType:"gauge" metricSource:"runtime"`
	HeapReleased  float64 `metricType:"gauge" metricSource:"runtime"`
	HeapSys       float64 `metricType:"gauge" metricSource:"runtime"`
	LastGC        float64 `metricType:"gauge" metricSource:"runtime"`
	Lookups       float64 `metricType:"gauge" metricSource:"runtime"`
	MCacheInuse   float64 `metricType:"gauge" metricSource:"runtime"`
	MCacheSys     float64 `metricType:"gauge" metricSource:"runtime"`
	MSpanInuse    float64 `metricType:"gauge" metricSource:"runtime"`
	MSpanSys      float64 `metricType:"gauge" metricSource:"runtime"`
	Mallocs       float64 `metricType:"gauge" metricSource:"runtime"`
	NextGC        float64 `metricType:"gauge" metricSource:"runtime"`
	NumForcedGC   float64 `metricType:"gauge" metricSource:"runtime"`
	NumGC         float64 `metricType:"gauge" metricSource:"runtime"`
	OtherSys      float64 `metricType:"gauge" metricSource:"runtime"`
	PauseTotalNs  float64 `metricType:"gauge" metricSource:"runtime"`
	StackInuse    float64 `metricType:"gauge" metricSource:"runtime"`
	StackSys      float64 `metricType:"gauge" metricSource:"runtime"`
	Sys           float64 `metricType:"gauge" metricSource:"runtime"`
	TotalAlloc    float64 `metricType:"gauge" metricSource:"runtime"`
	RandomValue   float64 `metricType:"gauge" metricSource:"custom"`
	PollCount     int64   `metricType:"counter" metricSource:"custom"`
}

func NewStats() *Stats {
	return &Stats{}
}

func (s Stats) StatsToMetrics() []model.Metric {
	var metrics []model.Metric

	// Get reflection value of stats
	t := reflect.TypeOf(s)

	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		fieldName := fieldType.Name

		metric := model.Metric{
			ID: domain.MetricName(fieldName),
		}

		metricType := domain.MetricType(fieldType.Tag.Get("metricType"))
		metricSource := domain.MetricSource(fieldType.Tag.Get("metricSource"))

		// Check the field type to assign Counter or Value
		switch metricType {
		case domain.Gauge:
			metric.MType = domain.Gauge
		case domain.Counter:
			metric.MType = domain.Counter
		default:

			panic("unhandled metric type")
		}

		switch metricSource {
		case domain.Runtime:
			metric.MSource = domain.Runtime
		case domain.Custom:
			metric.MSource = domain.Custom
		default:

			panic("unhandled metric source")
		}

		metrics = append(metrics, metric)
	}

	return metrics
}
