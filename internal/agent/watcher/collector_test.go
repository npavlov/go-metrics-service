package watcher

import (
	"context"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/stats"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestCollector_UpdateMetrics(t *testing.T) {
	st := stats.Stats{}
	metrics := st.StatsToMetrics()
	mux := sync.RWMutex{}
	cfg := &config.Config{
		Address: "",
	}
	var collector = NewMetricCollector(&metrics, &mux, cfg)

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

// TestStartCollector tests the StartCollector method of MetricCollector
func TestStartCollector(t *testing.T) {
	//setup
	st := stats.Stats{}
	metrics := st.StatsToMetrics()
	var mu sync.RWMutex
	cfg := &config.Config{PollInterval: 1} // Poll every second

	// Create an instance of MetricCollector
	mc := NewMetricCollector(&metrics, &mu, cfg)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure that we cancel context after test completion

	var wg sync.WaitGroup
	wg.Add(1)

	// Run StartCollector in a goroutine
	go mc.StartCollector(ctx, &wg)

	// Sleep for a short period to allow the collector to run
	time.Sleep(3 * time.Second)

	// Cancel the context to stop the collector
	cancel()

	// Wait for the goroutine to finish
	wg.Wait()

	assert.NotNil(t, metrics) // Ensure metrics are not nil
	for _, val := range metrics {
		if val.ID == domain.PollCount {
			assert.Greater(t, *val.Counter, int64(0))
		}
	}

	assert.Equal(t, context.Canceled, ctx.Err())
}
