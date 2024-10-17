package storage

import (
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/pkg/errors"
	"strconv"
	"sync"
)

const (
	errInvalidGauge   = "invalid gauge value"
	errInvalidCounter = "invalid counter value"
	errUnknownMetric  = "unknown metric type"
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
