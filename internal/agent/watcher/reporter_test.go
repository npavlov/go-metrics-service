package watcher_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/utils"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

func TestMetricReporter_SendSingleMetric(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := zerolog.Nop()
	cfg := &config.Config{
		ReportIntervalDur: 100 * time.Millisecond,
		RateLimit:         1,
		UseBatch:          false,
	}

	// Mock server to simulate the Sender
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var receivedMetric db.Metric
		err := json.NewDecoder(request.Body).Decode(&receivedMetric)
		assert.NoError(t, err)

		assert.Equal(t, "metric1", receivedMetric.ID)

		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(receivedMetric)
	}))
	defer server.Close()

	cfg.Address = server.URL

	inputStream := make(chan []db.Metric, 1)
	inputStream <- []db.Metric{
		*db.NewMetric("metric1", "gauge", nil, float64Ptr(1.23)),
	}
	close(inputStream)

	reporter := watcher.NewMetricReporter(inputStream, cfg, &logger)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go reporter.StartReporter(ctx, wg)

	// Allow some time for the reporter to process
	time.Sleep(300 * time.Millisecond)
	cancel()
	wg.Wait()
}

func TestMetricReporter_SendBatchMetrics(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := zerolog.Nop()
	cfg := &config.Config{
		ReportIntervalDur: 100 * time.Millisecond,
		RateLimit:         1,
		UseBatch:          true,
	}

	// Mock server to simulate the Sender
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var receivedMetrics []db.Metric
		result, _ := utils.DecompressResult(request.Body)
		err := json.Unmarshal(result, &receivedMetrics)
		assert.NoError(t, err)

		assert.Len(t, receivedMetrics, 2)
		assert.Equal(t, domain.MetricName("metric1"), receivedMetrics[0].ID)
		assert.Equal(t, domain.MetricName("metric2"), receivedMetrics[1].ID)

		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(receivedMetrics)
	}))
	defer server.Close()

	cfg.Address = server.URL

	inputStream := make(chan []db.Metric, 1)
	inputStream <- []db.Metric{
		*db.NewMetric("metric1", "counter", int64Ptr(42), nil),
		*db.NewMetric("metric2", "gauge", nil, float64Ptr(3.14)),
	}

	reporter := watcher.NewMetricReporter(inputStream, cfg, &logger)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go reporter.StartReporter(ctx, wg)

	// Allow some time for the reporter to process
	time.Sleep(500 * time.Millisecond)
	cancel()
	wg.Wait()
	close(inputStream)
}

func TestMetricReporter_ErrorHandling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := zerolog.Nop()
	cfg := &config.Config{
		ReportIntervalDur: 100 * time.Millisecond,
		RateLimit:         1,
		UseBatch:          false,
	}

	// Mock server to simulate the Sender
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg.Address = server.URL

	inputStream := make(chan []db.Metric, 1)
	inputStream <- []db.Metric{
		*db.NewMetric("metric1", "gauge", nil, float64Ptr(1.23)),
	}
	close(inputStream)

	reporter := watcher.NewMetricReporter(inputStream, cfg, &logger)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go reporter.StartReporter(ctx, wg)

	// Allow some time for the reporter to process
	time.Sleep(300 * time.Millisecond)
	cancel()
	wg.Wait()
}

func TestMetricReporter_StopOnContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	logger := zerolog.Nop()
	cfg := &config.Config{
		ReportIntervalDur: 100 * time.Millisecond,
		RateLimit:         1,
		UseBatch:          false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var receivedMetric db.Metric
		result, _ := utils.DecompressResult(request.Body)
		err := json.Unmarshal(result, &receivedMetric)
		assert.NoError(t, err)

		assert.Equal(t, domain.MetricName("metric1"), receivedMetric.ID)
		assert.InDelta(t, 1.23, *receivedMetric.Value, 0.01)
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg.Address = server.URL

	inputStream := make(chan []db.Metric, 1)
	inputStream <- []db.Metric{
		*db.NewMetric("metric1", "gauge", nil, float64Ptr(1.23)),
	}

	reporter := watcher.NewMetricReporter(inputStream, cfg, &logger)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go reporter.StartReporter(ctx, wg)

	// Allow the reporter to start
	time.Sleep(400 * time.Millisecond)

	// Cancel the context and ensure it shuts down
	cancel()
	wg.Wait()

	assert.NotPanics(t, func() { close(inputStream) }, "inputStream should be safe to close after reporter shutdown")
}
