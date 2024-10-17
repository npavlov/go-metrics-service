package storage

import (
	"github.com/npavlov/go-metrics-service/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	ms := NewMemStorage()

	assert.NotNil(t, ms, "MemStorage instance should not be nil")
	assert.Empty(t, ms.gauges, "Gauges map should be empty initially")
	assert.Empty(t, ms.counters, "Counters map should be empty initially")
}

func TestUpdateGauge(t *testing.T) {
	ms := NewMemStorage()
	metricName := domain.MetricName("test_gauge")

	// Update gauge
	err := ms.UpdateMetric("gauge", metricName, "12.34")
	assert.Nil(t, err)

	// Validate the update
	value, exists := ms.GetGauge(metricName)
	assert.True(t, exists, "Gauge should exist")
	assert.Equal(t, 12.34, value, "Gauge value should be updated to 12.34")
}

func TestUpdateCounter(t *testing.T) {
	ms := NewMemStorage()
	metricName := domain.MetricName("test_counter")

	// Update counter
	err := ms.UpdateMetric("counter", metricName, "10")
	assert.Nil(t, err)
	// Validate the update
	value, exists := ms.GetCounter(metricName)
	assert.True(t, exists, "Counter should exist")
	assert.Equal(t, int64(10), value, "Counter value should be updated to 10")

	// Increment the counter
	err = ms.UpdateMetric("counter", metricName, "5")
	assert.Nil(t, err)
	value, _ = ms.GetCounter(metricName)
	assert.Equal(t, int64(15), value, "Counter value should be updated to 15 after increment")
}

func TestGetGauge(t *testing.T) {
	ms := NewMemStorage()
	metricName := domain.MetricName("non_existent_gauge")

	// Try to retrieve a non-existent gauge
	_, exists := ms.GetGauge(metricName)
	assert.False(t, exists, "Non-existent gauge should not be found")
}

func TestGetCounter(t *testing.T) {
	ms := NewMemStorage()
	metricName := domain.MetricName("non_existent_counter")

	// Try to retrieve a non-existent counter
	_, exists := ms.GetCounter(metricName)
	assert.False(t, exists, "Non-existent counter should not be found")
}

func TestGetGauges(t *testing.T) {
	ms := NewMemStorage()
	metricName := domain.MetricName("test_gauge")

	// Add a gauge
	err := ms.UpdateMetric("gauge", metricName, "12.34")
	assert.Nil(t, err)

	// Retrieve all gauges
	gauges := ms.GetGauges()
	assert.Contains(t, gauges, metricName, "Gauges map should contain the updated gauge")
	assert.Equal(t, 12.34, gauges[metricName], "Gauge value should be 12.34 in the map")
}

func TestGetCounters(t *testing.T) {
	ms := &MemStorage{
		gauges: map[domain.MetricName]float64{
			domain.MetricName("test_gauge"): 12.34,
			domain.MetricName("next_gauge"): 0.00000004,
		},
		counters: map[domain.MetricName]int64{
			domain.MetricName("test_counter"): 10,
			domain.MetricName("next_counter"): 9999999999,
		},
	}

	assert.Equal(t, ms.gauges, ms.GetGauges())
	assert.Equal(t, ms.counters, ms.GetCounters())

	metricName := domain.MetricName("test_counter")

	// Add a counter, initial value was 10
	err := ms.UpdateMetric("counter", metricName, "10")
	assert.Nil(t, err)

	// Retrieve all counters
	counters := ms.GetCounters()
	assert.Contains(t, counters, metricName, "Counters map should contain the updated counter")
	assert.Equal(t, int64(20), counters[metricName], "Counter value should be 10 in the map")
}
