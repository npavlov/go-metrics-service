package watcher

import (
	"github.com/npavlov/go-metrics-service/internal/agent/stats"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestCollector_UpdateMetrics(t *testing.T) {
	st := stats.Stats{}
	metrics := st.StatsToMetrics()
	mux := sync.RWMutex{}
	var collector = NewMetricCollector(&metrics, &mux)

	// Call the method to test
	collector.UpdateMetrics()

	for _, metric := range metrics {
		value, ok := metric.GetValue()

		assert.True(t, ok)
		assert.NotNil(t, value)

		if metric.ID == domain.PollCount {
			assert.Equal(t, "1", value)
		}
	}

	collector.UpdateMetrics()

	for _, metric := range metrics {
		value, ok := metric.GetValue()

		assert.True(t, ok)
		assert.NotNil(t, value)

		if metric.ID == domain.PollCount {
			assert.Equal(t, "2", value)
		}
	}
}
