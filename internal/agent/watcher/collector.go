package watcher

import (
	"context"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	errNoSuchField     = "no such field in memStats"
	errUnsupportedKind = "cannot convert field to float64, unsupported kind: %s"
)

// Collector interface defines the contract for updating watcher.
type Collector interface {
	UpdateMetrics()
	StartCollector(ctx context.Context, wg *sync.WaitGroup)
}

// MetricCollector implements the Collector interface.
type MetricCollector struct {
	metrics *[]model.Metric
	mux     *sync.RWMutex
	cfg     *config.Config
	l       *zerolog.Logger
}

// NewMetricCollector creates a new instance of MetricCollector.
func NewMetricCollector(m *[]model.Metric, mux *sync.RWMutex, cfg *config.Config, l *zerolog.Logger) *MetricCollector {
	return &MetricCollector{
		metrics: m,
		mux:     mux,
		cfg:     cfg,
		l:       l,
	}
}

func (mc *MetricCollector) StartCollector(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			mc.l.Info().Msg("Stopping watcher collection")

			return
		default:
			// Add your watcher collection logic here
			mc.UpdateMetrics()
			time.Sleep(time.Duration(mc.cfg.PollInterval) * time.Second)
		}
	}
}

// UpdateMetrics updates runtime watcher using the runtime package.
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
				mc.l.Error().Err(err).Str("Metric Id", string(metric.ID)).Msg("can't transform field to Float64")

				return
			}
			metric.SetValue(nil, &value)
		case domain.Custom:
			if metric.ID == domain.PollCount {
				val := int64(1)
				metric.SetValue(&val, nil)
			}
			if metric.ID == domain.RandomValue {
				//nolint:gosec
				val := rand.Float64()
				metric.SetValue(nil, &val)
			}
		}
	}
}

func (mc *MetricCollector) getFieldAsFloat64(value reflect.Value) (float64, error) {
	if !value.IsValid() {
		return 0, errors.New(errNoSuchField)
	}

	// Switch based on the field type and convert it to float64
	//nolint:exhaustive
	switch value.Kind() {
	case reflect.Uint64:
		return float64(value.Uint()), nil
	case reflect.Uint32:
		return float64(value.Uint()), nil
	case reflect.Float64:
		return value.Float(), nil
	default:
		return 0, errors.New(errUnsupportedKind)
	}
}
