package storage

import (
	. "github.com/npavlov/go-metrics-service/internal/metric-types"
	"sync"
)

type Repository interface {
	UpdateGauge(name MetricName, value float64)
	UpdateCounter(name MetricName, value int64)
	IncCounter(name MetricName)
	GetGauge(name MetricName) (float64, bool)
	GetCounter(name MetricName) (int64, bool)
	GetGauges() map[MetricName]float64
	GetCounters() map[MetricName]int64
}

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[MetricName]float64
	counters map[MetricName]int64
}

// NewMemStorage - конструктор для MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[MetricName]float64),
		counters: make(map[MetricName]int64),
	}
}

func (ms *MemStorage) GetGauges() map[MetricName]float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return cloneMap(ms.gauges)
}

func (ms *MemStorage) GetCounters() map[MetricName]int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return cloneMap(ms.counters)
}

// UpdateGauge - обновление значения gauge
func (ms *MemStorage) UpdateGauge(name MetricName, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.gauges[name] = value
}

// UpdateCounter - обновление значения counter
func (ms *MemStorage) UpdateCounter(name MetricName, value int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.counters[name] += value
}

// GetGauge - retrieves the value of a gauge
func (ms *MemStorage) GetGauge(name MetricName) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.gauges[name]
	return value, exists
}

// GetCounter - retrieves the value of a counter
func (ms *MemStorage) GetCounter(name MetricName) (int64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.counters[name]
	return value, exists
}

func (ms *MemStorage) IncCounter(name MetricName) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.counters[name]++
}

// Generic function to clone a map with either int64 or float64 values
func cloneMap[K comparable, V int64 | float64](original map[K]V) map[K]V {
	cloned := make(map[K]V)
	for key, value := range original {
		cloned[key] = value
	}
	return cloned
}
