package watcher

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/model"
	"github.com/npavlov/go-metrics-service/internal/agent/stats"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
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
	stats         []model.Metric
	cfg           *config.Config
	log           *zerolog.Logger
	metricsStream chan []db.Metric
}

// NewMetricCollector creates a new instance of MetricCollector.
func NewMetricCollector(metricStream chan []db.Metric, cfg *config.Config, l *zerolog.Logger) *MetricCollector {
	return &MetricCollector{
		stats:         stats.NewStats().StatsToMetrics(),
		cfg:           cfg,
		log:           l,
		metricsStream: metricStream,
	}
}

func (mc *MetricCollector) StartCollector(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(mc.cfg.PollIntervalDur)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			mc.log.Info().Msg("Stopping watcher collection")

			return
		case <-ticker.C:
			// Add your watcher collection logic here
			mc.UpdateMetrics()

			mc.log.Info().Msg("Metrics updated")
		}
	}
}

// UpdateMetrics updates runtime watcher using the runtime package.
func (mc *MetricCollector) UpdateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	rMemStats := reflect.ValueOf(memStats)

	memoryStat, err := mem.VirtualMemory()
	if err != nil {
		mc.log.Error().Err(err).Msg("Failed to retrieve memory stats")

		return
	}
	rGopMemStats := reflect.ValueOf(*memoryStat)

	timesStat, err := cpu.Times(true)
	if err != nil {
		mc.log.Error().Err(err).Msg("Failed to retrieve CPU stats")

		return
	}

	updatedMetrics := make([]db.Metric, 0)
	for _, metric := range mc.stats {
		var newMetric []db.Metric
		switch metric.MSource {
		case domain.Runtime:
			newMetric = mc.processRuntimeMetric(metric, rMemStats)
		case domain.Custom:
			newMetric = mc.processCustomMetric(metric)
		case domain.GopsMem:
			newMetric = mc.processGopsMemMetric(metric, rGopMemStats)
		case domain.GopsCPU:
			newMetric = mc.processGopsCPUMetric(metric, timesStat)
		default:
			mc.log.Warn().Str("Metric Id", string(metric.ID)).Msg("Unhandled metric source")
		}

		updatedMetrics = append(updatedMetrics, newMetric...)
	}

	mc.sendMetrics(updatedMetrics)
}

func (mc *MetricCollector) processRuntimeMetric(metric model.Metric, rMemStats reflect.Value) []db.Metric {
	rValue := rMemStats.FieldByName(string(metric.ID))
	value, err := mc.getFieldAsFloat64(rValue)
	if err != nil {
		mc.log.Error().Err(err).Str("Metric Id", string(metric.ID)).Msg("Failed to transform runtime field to float64")

		return nil
	}

	newMetric := db.NewMetric(metric.ID, metric.MType, nil, &value)

	return []db.Metric{*newMetric}
}

func (mc *MetricCollector) processCustomMetric(metric model.Metric) []db.Metric {
	newMetric := db.NewMetric(metric.ID, metric.MType, nil, nil)

	//nolint:exhaustive
	switch metric.ID {
	case domain.PollCount:
		val := int64(1)
		newMetric.SetValue(&val, nil)
	case domain.RandomValue:
		//nolint:gosec
		val := rand.Float64()
		newMetric.SetValue(nil, &val)
	}

	return []db.Metric{*newMetric}
}

func (mc *MetricCollector) processGopsMemMetric(metric model.Metric, rGopMemStats reflect.Value) []db.Metric {
	rValue := rGopMemStats.FieldByName(string(metric.MAlias))
	value, err := mc.getFieldAsFloat64(rValue)
	if err != nil {
		mc.log.Error().Err(err).Str("Metric Id", string(metric.ID)).Msg("Failed to transform gops memory field to float64")

		return nil
	}

	newMetric := db.NewMetric(metric.ID, metric.MType, nil, &value)

	return []db.Metric{*newMetric}
}

func (mc *MetricCollector) processGopsCPUMetric(metric model.Metric, timesStat []cpu.TimesStat) []db.Metric {
	metrics := make([]db.Metric, 0)
	for i, cpuStat := range timesStat {
		newID := fmt.Sprintf("%s%d", string(metric.ID), i)
		newMetric := db.NewMetric(domain.MetricName(newID), metric.MType, nil, nil)

		rGopTimeStats := reflect.ValueOf(cpuStat)
		rValue := rGopTimeStats.FieldByName(string(metric.MAlias))
		value, err := mc.getFieldAsFloat64(rValue)
		if err != nil {
			mc.log.Error().Err(err).Str("Metric Id", string(metric.ID)).Msg("Failed to transform CPU stat field to float64")

			continue
		}

		newMetric.SetValue(nil, &value)
		metrics = append(metrics, *newMetric)
	}

	return metrics
}

func (mc *MetricCollector) sendMetrics(metrics []db.Metric) {
	go func() {
		mc.metricsStream <- metrics
	}()
}

func (mc *MetricCollector) getFieldAsFloat64(value reflect.Value) (float64, error) {
	if !value.IsValid() {
		return 0, errors.New(errNoSuchField)
	}

	// Switch based on the field type and convert it to float64
	//nolint:exhaustive // Only handling specific types we expect
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
