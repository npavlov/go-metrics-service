package storage

import "sync"

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type MemStorage struct {
	mu       sync.RWMutex
	Gauges   map[string]float64
	Counters map[string]int64
}

// NewMemStorage - конструктор для MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

// UpdateGauge - обновление значения gauge
func (ms *MemStorage) UpdateGauge(name string, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Gauges[name] = value
}

// UpdateCounter - обновление значения counter
func (ms *MemStorage) UpdateCounter(name string, value int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Counters[name] += value
}
