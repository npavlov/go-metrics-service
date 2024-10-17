package storage_test

import (
	"testing"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage(t *testing.T) {
	t.Parallel()

	// Test Update and Get for Gauge
	t.Run("Update and Get Gauge", func(t *testing.T) {
		t.Parallel()
		st := storage.NewMemStorage()

		gaugeName := domain.MetricName("test.gauge")
		err := st.UpdateMetric(domain.Gauge, gaugeName, "10.5")
		require.NoError(t, err)

		value, exists := st.GetGauge(gaugeName)
		assert.True(t, exists)
		assert.InDelta(t, 10.5, value, 0000.1)

		// Update again
		err = st.UpdateMetric(domain.Gauge, gaugeName, "20.5")
		require.NoError(t, err)

		value, exists = st.GetGauge(gaugeName)
		assert.True(t, exists)
		assert.InDelta(t, 20.5, value, 0000.1)
	})

	// Test Update and Get for Counter
	t.Run("Update and Get Counter", func(t *testing.T) {
		t.Parallel()

		st := storage.NewMemStorage()

		counterName := domain.MetricName("test.counter")
		err := st.UpdateMetric(domain.Counter, counterName, "5")
		require.NoError(t, err)

		value, exists := st.GetCounter(counterName)
		assert.True(t, exists)
		assert.Equal(t, int64(5), value)

		// Increment counter
		err = st.UpdateMetric(domain.Counter, counterName, "3")
		require.NoError(t, err)

		value, exists = st.GetCounter(counterName)
		assert.True(t, exists)
		assert.Equal(t, int64(8), value)
	})

	// Test error for invalid gauge value
	t.Run("Invalid Gauge Value", func(t *testing.T) {
		t.Parallel()

		st := storage.NewMemStorage()

		err := st.UpdateMetric(domain.Gauge, "test.gauge", "invalid")
		require.Error(t, err)
		assert.Equal(t, "invalid gauge value: strconv.ParseFloat: parsing \"invalid\": invalid syntax", err.Error())
	})

	// Test error for invalid counter value
	t.Run("Invalid Counter Value", func(t *testing.T) {
		t.Parallel()

		st := storage.NewMemStorage()

		err := st.UpdateMetric(domain.Counter, "test.counter", "invalid")
		require.Error(t, err)
		assert.Equal(t, "invalid counter value: strconv.ParseInt: parsing \"invalid\": invalid syntax", err.Error())
	})

	// Test GetMetricModel for Gauge
	t.Run("Get Metric Model for Gauge", func(t *testing.T) {
		t.Parallel()

		st := storage.NewMemStorage()

		gaugeMetric := &model.Metric{
			ID:    "test.gauge",
			MType: domain.Gauge,
			Value: float64Ptr(15.5),
		}
		err := st.UpdateMetricModel(gaugeMetric)
		require.NoError(t, err)

		retrievedMetric, err := st.GetMetricModel(gaugeMetric)
		require.NoError(t, err)
		assert.Equal(t, gaugeMetric.ID, retrievedMetric.ID)
		assert.InDelta(t, *gaugeMetric.Value, *retrievedMetric.Value, 0000.1)

		gauges := st.GetGauges()
		assert.Len(t, gauges, 1)
	})

	// Test GetMetricModel for Counter
	t.Run("Get Metric Model for Counter", func(t *testing.T) {
		t.Parallel()

		st := storage.NewMemStorage()

		counterMetric := &model.Metric{
			ID:    "test.counter",
			MType: domain.Counter,
			Delta: int64Ptr(10),
		}
		err := st.UpdateMetricModel(counterMetric)
		require.NoError(t, err)

		retrievedMetric, err := st.GetMetricModel(counterMetric)
		require.NoError(t, err)
		assert.Equal(t, counterMetric.ID, retrievedMetric.ID)
		assert.Equal(t, *counterMetric.Delta, *retrievedMetric.Delta)

		counters := st.GetCounters()
		assert.Len(t, counters, 1)
	})

	// Test error when no value is provided
	t.Run("No Value Error", func(t *testing.T) {
		t.Parallel()

		st := storage.NewMemStorage()

		metric := &model.Metric{
			ID:    "test.no.value",
			MType: domain.Gauge,
		}
		err := st.UpdateMetricModel(metric)
		require.Error(t, err)
		assert.Equal(t, "no value provided", err.Error())
	})

	// Test unknown metric type
	t.Run("Unknown Metric Type", func(t *testing.T) {
		t.Parallel()

		st := storage.NewMemStorage()

		metric := &model.Metric{
			ID:    "test.unknown",
			MType: "unknown",
			Value: float64Ptr(15.5),
		}
		err := st.UpdateMetricModel(metric)
		require.Error(t, err)
		assert.Equal(t, "unknown metric type", err.Error())
	})
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
