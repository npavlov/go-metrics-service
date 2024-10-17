package storage_test

import (
	"testing"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemStorage(t *testing.T) {
	t.Parallel()

	ms := storage.NewMemStorage()

	assert.NotNil(t, ms, "MemStorage instance should not be nil")
	assert.Empty(t, ms.GetGauges(), "Gauges map should be empty initially")
	assert.Empty(t, ms.GetCounters(), "Counters map should be empty initially")
}

func TestUpdateGauge(t *testing.T) {
	t.Parallel()

	ms := storage.NewMemStorage()
	metricName := domain.MetricName("test_gauge")

	// Update gauge
	err := ms.UpdateMetric("gauge", metricName, "12.34")
	require.NoError(t, err)

	// Validate the update
	value, exists := ms.GetGauge(metricName)
	assert.True(t, exists, "Gauge should exist")
	assert.InDelta(t, 12.34, value, 0.000001, "Gauge value should be updated to 12.34")
}

func TestUpdateCounter(t *testing.T) {
	t.Parallel()

	ms := storage.NewMemStorage()
	metricName := domain.MetricName("test_counter")

	// Update counter
	err := ms.UpdateMetric("counter", metricName, "10")
	require.NoError(t, err)
	// Validate the update
	value, exists := ms.GetCounter(metricName)
	assert.True(t, exists, "Delta should exist")
	assert.Equal(t, int64(10), value, "Delta value should be updated to 10")

	// Increment the counter
	err = ms.UpdateMetric("counter", metricName, "5")
	require.NoError(t, err)
	value, _ = ms.GetCounter(metricName)
	assert.Equal(t, int64(15), value, "Delta value should be updated to 15 after increment")
}

func TestGetGauge(t *testing.T) {
	t.Parallel()

	ms := storage.NewMemStorage()
	metricName := domain.MetricName("non_existent_gauge")

	// Try to retrieve a non-existent gauge
	_, exists := ms.GetGauge(metricName)
	assert.False(t, exists, "Non-existent gauge should not be found")
}

func TestGetCounter(t *testing.T) {
	t.Parallel()

	ms := storage.NewMemStorage()
	metricName := domain.MetricName("non_existent_counter")

	// Try to retrieve a non-existent counter
	_, exists := ms.GetCounter(metricName)
	assert.False(t, exists, "Non-existent counter should not be found")
}

func TestGetGauges(t *testing.T) {
	t.Parallel()

	ms := storage.NewMemStorage()
	metricName := domain.MetricName("test_gauge")

	// Add a gauge
	err := ms.UpdateMetric("gauge", metricName, "12.34")
	require.NoError(t, err)

	// Retrieve all gauges
	gauges := ms.GetGauges()
	assert.Contains(t, gauges, metricName, "Gauges map should contain the updated gauge")
	assert.InDelta(t, 12.34, gauges[metricName], 0.000001, "Gauge value should be 12.34 in the map")
}

func TestGetCounters(t *testing.T) {
	t.Parallel()

	ms := storage.NewMemStorage()
	err := ms.UpdateMetric("gauge", "test_gauge", "12.34")
	require.NoError(t, err)
	err = ms.UpdateMetric("gauge", "next_gauge", "0.00000004")
	require.NoError(t, err)
	err = ms.UpdateMetric("counter", "test_counter", "10")
	require.NoError(t, err)
	err = ms.UpdateMetric("counter", "next_counter", "9999999999")
	require.NoError(t, err)
	gauges := ms.GetGauges()
	assert.Contains(t, gauges, domain.MetricName("next_gauge"), "Gauges map should contain the updated gauge")
	counters := ms.GetCounters()
	assert.Contains(t, counters, domain.MetricName("next_counter"), "Counters map should contain the updated counter")

	metricName := domain.MetricName("test_counter")

	// Add a counter, initial value was 10
	err = ms.UpdateMetric("counter", metricName, "10")
	require.NoError(t, err)

	// Retrieve all counters
	counters = ms.GetCounters()
	assert.Contains(t, counters, metricName, "Counters map should contain the updated counter")
	assert.Equal(t, int64(20), counters[metricName], "Delta value should be 10 in the map")
}
