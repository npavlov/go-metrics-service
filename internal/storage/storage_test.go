package storage

import (
	"testing"

	types "github.com/npavlov/go-metrics-service/internal/agent/metrictypes"
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
	metricName := types.MetricName("test_gauge")

	// Update gauge
	ms.UpdateGauge(metricName, 12.34)

	// Validate the update
	value, exists := ms.GetGauge(metricName)
	assert.True(t, exists, "Gauge should exist")
	assert.Equal(t, 12.34, value, "Gauge value should be updated to 12.34")
}

func TestUpdateCounter(t *testing.T) {
	ms := NewMemStorage()
	metricName := types.MetricName("test_counter")

	// Update counter
	ms.UpdateCounter(metricName, 10)

	// Validate the update
	value, exists := ms.GetCounter(metricName)
	assert.True(t, exists, "Counter should exist")
	assert.Equal(t, int64(10), value, "Counter value should be updated to 10")

	// Increment the counter
	ms.UpdateCounter(metricName, 5)
	value, _ = ms.GetCounter(metricName)
	assert.Equal(t, int64(15), value, "Counter value should be updated to 15 after increment")
}

func TestIncCounter(t *testing.T) {
	ms := NewMemStorage()
	metricName := types.MetricName("test_inc_counter")

	// Increment the counter
	ms.IncCounter(metricName)

	// Validate the increment
	value, exists := ms.GetCounter(metricName)
	assert.True(t, exists, "Counter should exist after increment")
	assert.Equal(t, int64(1), value, "Counter value should be incremented by 1")

	// Increment again
	ms.IncCounter(metricName)
	value, _ = ms.GetCounter(metricName)
	assert.Equal(t, int64(2), value, "Counter value should be incremented to 2 after second increment")
}

func TestGetGauge(t *testing.T) {
	ms := NewMemStorage()
	metricName := types.MetricName("non_existent_gauge")

	// Try to retrieve a non-existent gauge
	_, exists := ms.GetGauge(metricName)
	assert.False(t, exists, "Non-existent gauge should not be found")
}

func TestGetCounter(t *testing.T) {
	ms := NewMemStorage()
	metricName := types.MetricName("non_existent_counter")

	// Try to retrieve a non-existent counter
	_, exists := ms.GetCounter(metricName)
	assert.False(t, exists, "Non-existent counter should not be found")
}

func TestGetGauges(t *testing.T) {
	ms := NewMemStorage()
	metricName := types.MetricName("test_gauge")

	// Add a gauge
	ms.UpdateGauge(metricName, 12.34)

	// Retrieve all gauges
	gauges := ms.GetGauges()
	assert.Contains(t, gauges, metricName, "Gauges map should contain the updated gauge")
	assert.Equal(t, 12.34, gauges[metricName], "Gauge value should be 12.34 in the map")
}

func TestGetCounters(t *testing.T) {
	ms := &MemStorage{
		gauges: map[types.MetricName]float64{
			types.MetricName("test_gauge"): 12.34,
			types.MetricName("next_gauge"): 0.00000004,
		},
		counters: map[types.MetricName]int64{
			types.MetricName("test_counter"): 10,
			types.MetricName("next_counter"): 9999999999,
		},
	}

	assert.Equal(t, ms.gauges, ms.GetGauges())
	assert.Equal(t, ms.counters, ms.GetCounters())

	metricName := types.MetricName("test_counter")

	// Add a counter, initial value was 10
	ms.UpdateCounter(metricName, 10)

	// Retrieve all counters
	counters := ms.GetCounters()
	assert.Contains(t, counters, metricName, "Counters map should contain the updated counter")
	assert.Equal(t, int64(20), counters[metricName], "Counter value should be 10 in the map")
}