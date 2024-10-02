package storage

import (
	"github.com/npavlov/go-metrics-service/internal/types"
	"github.com/pkg/errors"
	"strconv"
	"sync"
)

const (
	errInvalidGauge   = "invalid gauge value"
	errInvalidCounter = "invalid counter value"
	errUnknownMetric  = "unknown metric type"
)

type Repository interface {
	GetGauge(name types.MetricName) (float64, bool)
	GetCounter(name types.MetricName) (int64, bool)
	GetGauges() map[types.MetricName]float64
	GetCounters() map[types.MetricName]int64
	UpdateMetric(metricType types.MetricType, metricName types.MetricName, metricValue string) error
}

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[types.MetricName]float64
	counters map[types.MetricName]int64
}

// NewMemStorage - constructor for MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[types.MetricName]float64),
		counters: make(map[types.MetricName]int64),
	}
}

func (ms *MemStorage) GetGauges() map[types.MetricName]float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return cloneMap(ms.gauges)
}

func (ms *MemStorage) GetCounters() map[types.MetricName]int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return cloneMap(ms.counters)
}

// GetGauge - retrieves the value of a gauge
func (ms *MemStorage) GetGauge(name types.MetricName) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.gauges[name]
	return value, exists
}

// GetCounter - retrieves the value of a counter
func (ms *MemStorage) GetCounter(name types.MetricName) (int64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.counters[name]
	return value, exists
}

// Generic function to clone a map with either int64 or float64 values
func cloneMap[K comparable, V int64 | float64](original map[K]V) map[K]V {
	cloned := make(map[K]V)
	for key, value := range original {
		cloned[key] = value
	}
	return cloned
}

func (ms *MemStorage) UpdateMetric(metricType types.MetricType, metricName types.MetricName, metricValue string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	switch metricType {
	case types.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return errors.Wrap(err, errInvalidGauge)
		}
		ms.gauges[metricName] = value
	case types.Counter:
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
