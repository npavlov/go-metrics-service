package watcher_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/stats"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/domain"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestCollector_UpdateMetrics(t *testing.T) {
	t.Parallel()

	st := stats.Stats{}
	metrics := st.StatsToMetrics()
	mux := sync.RWMutex{}
	cfg := &config.Config{
		Address:        "",
		ReportInterval: 1,
		PollInterval:   1,
	}
	l := testutils.GetTLogger()
	newConfig := config.NewConfigBuilder(l).FromObj(cfg).Build()
	collector := watcher.NewMetricCollector(&metrics, &mux, newConfig, l)

	// Call the method to test
	collector.UpdateMetrics()

	for _, metric := range metrics {
		if metric.ID == domain.PollCount {
			delta := *metric.Delta
			assert.Equal(t, int64(1), delta)
		}
	}

	collector.UpdateMetrics()

	for _, metric := range metrics {
		if metric.ID == domain.PollCount {
			delta := *metric.Delta
			assert.Equal(t, int64(2), delta)
		}
	}
}

// TestStartCollector tests the StartCollector method of MetricCollector.
func TestStartCollector(t *testing.T) {
	t.Parallel()

	// setup
	st := stats.Stats{}
	metrics := st.StatsToMetrics()
	var mu sync.RWMutex
	cfg := &config.Config{PollInterval: 1, Address: "", ReportInterval: 1} // Poll every second
	l := testutils.GetTLogger()
	newConfig := config.NewConfigBuilder(l).FromObj(cfg).Build()

	// Create an instance of MetricCollector
	mc := watcher.NewMetricCollector(&metrics, &mu, newConfig, l)

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
			assert.Positive(t, *val.Delta)
		}
	}

	assert.Equal(t, context.Canceled, ctx.Err())
}
