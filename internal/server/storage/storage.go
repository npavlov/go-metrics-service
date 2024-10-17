package storage

import (
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/pkg/errors"
	"strconv"
	"sync"
)

const (
	errInvalidGauge   = "invalid gauge value"
	errInvalidCounter = "invalid counter value"
	errUnknownMetric  = "unknown metric type"
	errNoValue        = "no value provided"
)

type Number interface {
	int64 | float64
}

type Repository interface {
	GetGauge(name domain.MetricName) (float64, bool)
	GetCounter(name domain.MetricName) (int64, bool)
	GetGauges() map[domain.MetricName]float64
	GetCounters() map[domain.MetricName]int64
	UpdateMetric(metricType domain.MetricType, metricName domain.MetricName, metricValue string) error
	UpdateMetricModel(metric *model.Metric) error
	GetMetricModel(metric *model.Metric) (*model.Metric, error)
}

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[domain.MetricName]float64
	counters map[domain.MetricName]int64
}

// NewMemStorage - constructor for MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[domain.MetricName]float64),
		counters: make(map[domain.MetricName]int64),
	}
}

func (ms *MemStorage) GetGauges() map[domain.MetricName]float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return cloneMap(ms.gauges)
}

func (ms *MemStorage) GetCounters() map[domain.MetricName]int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return cloneMap(ms.counters)
}

// GetGauge - retrieves the value of a gauge
func (ms *MemStorage) GetGauge(name domain.MetricName) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.gauges[name]
	return value, exists
}

// GetCounter - retrieves the value of a counter
func (ms *MemStorage) GetCounter(name domain.MetricName) (int64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.counters[name]
	return value, exists
}

// Generic function to clone a map with either int64 or float64 values
func cloneMap[K comparable, V Number](original map[K]V) map[K]V {
	cloned := make(map[K]V)
	for key, value := range original {
		cloned[key] = value
	}
	return cloned
}

func (ms *MemStorage) UpdateMetric(metricType domain.MetricType, metricName domain.MetricName, metricValue string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	switch metricType {
	case domain.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return errors.Wrap(err, errInvalidGauge)
		}
		ms.gauges[metricName] = value
	case domain.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return errors.Wrap(err, errInvalidCounter)
		}
		ms.counters[metricName] += value
	default:
		return errors.New(errUnknownMetric)
	}
	return nil
}

func (ms *MemStorage) UpdateMetricModel(metric *model.Metric) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if metric.Delta == nil && metric.Value == nil {
		return errors.New(errNoValue)
	}

	switch metric.MType {
	case domain.Gauge:
		ms.gauges[metric.ID] = *(metric).Value
	case domain.Counter:
		ms.counters[metric.ID] += *(metric).Delta
	default:
		return errors.New(errUnknownMetric)
	}

	return nil
}

func (ms *MemStorage) GetMetricModel(metric *model.Metric) (*model.Metric, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	switch metric.MType {
	case domain.Gauge:
		if val, exists := ms.gauges[metric.ID]; exists {
			return &model.Metric{ID: metric.ID, MType: metric.MType, Value: &val}, nil
		}
		return nil, errors.New(errInvalidGauge)

	case domain.Counter:
		if delta, exists := ms.counters[metric.ID]; exists {
			return &model.Metric{ID: metric.ID, MType: metric.MType, Delta: &delta}, nil
		}
		return nil, errors.New(errInvalidCounter)

	default:
		return nil, errors.New(errUnknownMetric)
	}
}
