package watcher_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/server/db"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/domain"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestCollector_UpdateMetrics(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Address:        "",
		ReportInterval: 10,
		PollInterval:   10,
	}
	l := testutils.GetTLogger()
	newConfig := config.NewConfigBuilder(l).FromObj(cfg).Build()
	metricsStream := make(chan []db.Metric, 1)
	collector := watcher.NewMetricCollector(metricsStream, newConfig, l)

	// Call the method to test
	collector.UpdateMetrics()

	metrics := <-metricsStream

	for _, metric := range metrics {
		if metric.ID == domain.PollCount {
			delta := *metric.Delta
			assert.Equal(t, int64(1), delta)
		}
	}

	collector.UpdateMetrics()

	metrics = <-metricsStream

	for _, metric := range metrics {
		if metric.ID == domain.PollCount {
			delta := *metric.Delta
			assert.Equal(t, int64(1), delta)
		}
	}
}

// TestStartCollector tests the StartCollector method of MetricCollector.
func TestStartCollector(t *testing.T) {
	t.Parallel()

	// setup
	cfg := &config.Config{PollInterval: 1, Address: "", ReportInterval: 1} // Poll every second
	logger := testutils.GetTLogger()
	newConfig := config.NewConfigBuilder(logger).FromObj(cfg).Build()

	metricsStream := make(chan []db.Metric, 10)

	// Create an instance of MetricCollector
	mc := watcher.NewMetricCollector(metricsStream, newConfig, logger)

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

	metrics := <-metricsStream
	assert.NotNil(t, metrics)
	for _, val := range metrics {
		if val.ID == domain.PollCount {
			assert.Positive(t, *val.Delta)
		}
	}

	assert.Equal(t, context.Canceled, ctx.Err())
}
