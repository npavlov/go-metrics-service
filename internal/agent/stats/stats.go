package stats

import (
	"reflect"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
)

type Stats struct {
	Alloc         float64 `metricSource:"runtime" metricType:"gauge"`
	BuckHashSys   float64 `metricSource:"runtime" metricType:"gauge"`
	Frees         float64 `metricSource:"runtime" metricType:"gauge"`
	GCCPUFraction float64 `metricSource:"runtime" metricType:"gauge"`
	GCSys         float64 `metricSource:"runtime" metricType:"gauge"`
	HeapAlloc     float64 `metricSource:"runtime" metricType:"gauge"`
	HeapIdle      float64 `metricSource:"runtime" metricType:"gauge"`
	HeapInuse     float64 `metricSource:"runtime" metricType:"gauge"`
	HeapObjects   float64 `metricSource:"runtime" metricType:"gauge"`
	HeapReleased  float64 `metricSource:"runtime" metricType:"gauge"`
	HeapSys       float64 `metricSource:"runtime" metricType:"gauge"`
	LastGC        float64 `metricSource:"runtime" metricType:"gauge"`
	Lookups       float64 `metricSource:"runtime" metricType:"gauge"`
	MCacheInuse   float64 `metricSource:"runtime" metricType:"gauge"`
	MCacheSys     float64 `metricSource:"runtime" metricType:"gauge"`
	MSpanInuse    float64 `metricSource:"runtime" metricType:"gauge"`
	MSpanSys      float64 `metricSource:"runtime" metricType:"gauge"`
	Mallocs       float64 `metricSource:"runtime" metricType:"gauge"`
	NextGC        float64 `metricSource:"runtime" metricType:"gauge"`
	NumForcedGC   float64 `metricSource:"runtime" metricType:"gauge"`
	NumGC         float64 `metricSource:"runtime" metricType:"gauge"`
	OtherSys      float64 `metricSource:"runtime" metricType:"gauge"`
	PauseTotalNs  float64 `metricSource:"runtime" metricType:"gauge"`
	StackInuse    float64 `metricSource:"runtime" metricType:"gauge"`
	StackSys      float64 `metricSource:"runtime" metricType:"gauge"`
	Sys           float64 `metricSource:"runtime" metricType:"gauge"`
	TotalAlloc    float64 `metricSource:"runtime" metricType:"gauge"`
	RandomValue   float64 `metricSource:"custom"  metricType:"gauge"`
	PollCount     int64   `metricSource:"custom"  metricType:"counter"`
}

func NewStats() *Stats {
	//exhaustruct:ignore
	return &Stats{}
}

func (s *Stats) StatsToMetrics() []model.Metric {
	// Get reflection value of stats
	t := reflect.TypeOf(*s)

	metrics := make([]model.Metric, t.NumField())

	for index := range t.NumField() {
		fieldType := t.Field(index)
		fieldName := fieldType.Name

		metric := model.Metric{
			ID:      domain.MetricName(fieldName),
			MType:   domain.MetricType(fieldType.Tag.Get("metricType")),
			MSource: domain.MetricSource(fieldType.Tag.Get("metricSource")),
			Value:   nil,
			Delta:   nil,
		}

		if metric.MType != domain.Gauge && metric.MType != domain.Counter {
			panic("unhandled metric type")
		}
		if metric.MSource != domain.Runtime && metric.MSource != domain.Custom {
			panic("unhandled metric source")
		}

		metrics[index] = metric
	}

	return metrics
}
