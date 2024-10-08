package watcher

import (
	"context"
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"time"
)

// Collector interface defines the contract for updating watcher
type Collector interface {
	UpdateMetrics()
	StartCollector(ctx context.Context, cfg *config.Config)
}

// MetricCollector implements the Collector interface
type MetricCollector struct {
	metrics *[]model.Metric
	mux     *sync.RWMutex
}

func (mc *MetricCollector) StartCollector(ctx context.Context, cfg *config.Config) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping watcher collection")
			return
		default:
			// Add your watcher collection logic here
			mc.UpdateMetrics()
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		}
	}
}

// NewMetricCollector creates a new instance of MetricCollector
func NewMetricCollector(metrics *[]model.Metric, mux *sync.RWMutex) *MetricCollector {
	return &MetricCollector{
		metrics: metrics,
		mux:     mux,
	}
}

// UpdateMetrics updates runtime watcher using the runtime package
func (mc *MetricCollector) UpdateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	rMemStats := reflect.ValueOf(memStats)

	mc.mux.Lock()
	defer mc.mux.Unlock()

	for i := range *mc.metrics {
		// Access by reference
		metric := &(*mc.metrics)[i]

		switch metric.MSource {
		case domain.Runtime:
			rValue := rMemStats.FieldByName(string(metric.ID))
			value, err := mc.getFieldAsFloat64(rValue)
			if err != nil {
				fmt.Printf("can't transform field to Float64 %s: %v", metric.ID, err)
				return
			}
			metric.SetValue(nil, &value)
			metric.GetValue()
		case domain.Custom:
			if metric.ID == domain.PollCount {
				val := int64(1)
				metric.SetValue(&val, nil)
			}
			if metric.ID == domain.RandomValue {
				val := rand.Float64()
				metric.SetValue(nil, &val)
			}
		}
	}
}

func (mc *MetricCollector) getFieldAsFloat64(value reflect.Value) (float64, error) {
	if !value.IsValid() {
		return 0, fmt.Errorf("no such field: %s in memStats", value)
	}

	// Switch based on the field type and convert it to float64
	switch value.Kind() {
	case reflect.Uint64:
		return float64(value.Uint()), nil
	case reflect.Uint32:
		return float64(value.Uint()), nil
	case reflect.Float64:
		return value.Float(), nil
	default:
		return 0, fmt.Errorf("cannot convert field to float64, unsupported kind: %s", value.Kind())
	}
}
