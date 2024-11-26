package stats

import (
	"reflect"
	"slices"

	"github.com/npavlov/go-metrics-service/internal/agent/model"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

type Stats struct {
	Alloc          float64   `metricSource:"runtime" metricType:"gauge"`
	BuckHashSys    float64   `metricSource:"runtime" metricType:"gauge"`
	Frees          float64   `metricSource:"runtime" metricType:"gauge"`
	GCCPUFraction  float64   `metricSource:"runtime" metricType:"gauge"`
	GCSys          float64   `metricSource:"runtime" metricType:"gauge"`
	HeapAlloc      float64   `metricSource:"runtime" metricType:"gauge"`
	HeapIdle       float64   `metricSource:"runtime" metricType:"gauge"`
	HeapInuse      float64   `metricSource:"runtime" metricType:"gauge"`
	HeapObjects    float64   `metricSource:"runtime" metricType:"gauge"`
	HeapReleased   float64   `metricSource:"runtime" metricType:"gauge"`
	HeapSys        float64   `metricSource:"runtime" metricType:"gauge"`
	LastGC         float64   `metricSource:"runtime" metricType:"gauge"`
	Lookups        float64   `metricSource:"runtime" metricType:"gauge"`
	MCacheInuse    float64   `metricSource:"runtime" metricType:"gauge"`
	MCacheSys      float64   `metricSource:"runtime" metricType:"gauge"`
	MSpanInuse     float64   `metricSource:"runtime" metricType:"gauge"`
	MSpanSys       float64   `metricSource:"runtime" metricType:"gauge"`
	Mallocs        float64   `metricSource:"runtime" metricType:"gauge"`
	NextGC         float64   `metricSource:"runtime" metricType:"gauge"`
	NumForcedGC    float64   `metricSource:"runtime" metricType:"gauge"`
	NumGC          float64   `metricSource:"runtime" metricType:"gauge"`
	OtherSys       float64   `metricSource:"runtime" metricType:"gauge"`
	PauseTotalNs   float64   `metricSource:"runtime" metricType:"gauge"`
	StackInuse     float64   `metricSource:"runtime" metricType:"gauge"`
	StackSys       float64   `metricSource:"runtime" metricType:"gauge"`
	Sys            float64   `metricSource:"runtime" metricType:"gauge"`
	TotalAlloc     float64   `metricSource:"runtime" metricType:"gauge"`
	RandomValue    float64   `metricSource:"custom"  metricType:"gauge"`
	PollCount      int64     `metricSource:"custom"  metricType:"counter"`
	TotalMemory    float64   `metricAlias:"Total"    metricSource:"gopsutil/mem" metricType:"gauge"`
	FreeMemory     float64   `metricAlias:"Free"     metricSource:"gopsutil/mem" metricType:"gauge"`
	CPUutilization []float64 `metricAlias:"System"   metricSource:"gopsutil/cpu" metricType:"gauge"`
}

func NewStats() *Stats {
	//exhaustruct:ignore
	return &Stats{}
}

func (s *Stats) getMetricAliases() []domain.MetricAlias {
	return []domain.MetricAlias{
		domain.MFree,
		domain.MTotal,
		domain.MSystem,
	}
}

func (s *Stats) getMetricSources() []domain.MetricSource {
	return []domain.MetricSource{
		domain.Custom,
		domain.Runtime,
		domain.GopsMem,
		domain.GopsCPU,
	}
}

func (s *Stats) StatsToMetrics() []model.Metric {
	// Get reflection value of stats
	t := reflect.TypeOf(*s)

	metrics := make([]model.Metric, t.NumField())

	for index := range t.NumField() {
		fieldType := t.Field(index)
		mID := domain.MetricName(fieldType.Name)
		mType := domain.MetricType(fieldType.Tag.Get("metricType"))

		metric := model.Metric{
			Metric:  *db.NewMetric(mID, mType, nil, nil),
			MSource: domain.MetricSource(fieldType.Tag.Get("metricSource")),
			MAlias:  domain.MetricAlias(fieldType.Tag.Get("metricAlias")),
		}

		if metric.MType != domain.Gauge && metric.MType != domain.Counter {
			panic("unhandled metric type")
		}
		if !slices.Contains(s.getMetricSources(), metric.MSource) {
			panic("unhandled metric source")
		}
		if len(metric.MAlias) > 0 && !slices.Contains(s.getMetricAliases(), metric.MAlias) {
			panic("unhandled metric alias")
		}

		metrics[index] = metric
	}

	return metrics
}
