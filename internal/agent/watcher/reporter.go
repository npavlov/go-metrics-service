package watcher

import (
	"context"
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/model"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

// Reporter interface defines the contract for sending watcher
type Reporter interface {
	SendMetrics(ctx context.Context)
	StartReporter(ctx context.Context, wg *sync.WaitGroup)
}

// MetricReporter implements the Reporter interface
type MetricReporter struct {
	metrics *[]model.Metric
	mux     *sync.RWMutex
	cfg     *config.Config
}

func NewMetricReporter(metrics *[]model.Metric, mutex *sync.RWMutex, cfg *config.Config) *MetricReporter {
	return &MetricReporter{
		metrics: metrics,
		mux:     mutex,
		cfg:     cfg,
	}
}

func (mr *MetricReporter) StartReporter(ctx context.Context, wg *sync.WaitGroup) {
	l := logger.Get()

	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			l.Info().Msg("Stopping watcher reporting")
			return
		default:
			// Add your watcher reporting logic here
			mr.SendMetrics(ctx)
			time.Sleep(time.Duration(mr.cfg.ReportInterval) * time.Second)
		}
	}
}

// SendMetrics sends the collected watcher to the server
func (mr *MetricReporter) SendMetrics(ctx context.Context) {
	mr.mux.Lock()
	defer mr.mux.Unlock()

	for _, metric := range *mr.metrics {
		val, found := metric.GetValue()
		if found {
			url := fmt.Sprintf("%s/update/%s/%s/%s", mr.cfg.Address, metric.MType, metric.ID, val)
			mr.sendPostRequest(url, ctx)
		}
	}
}

// sendPostRequest sends a POST request to the given URL
func (mr *MetricReporter) sendPostRequest(url string, ctx context.Context) {
	log := logger.Get()

	client := resty.New()
	resp, err := client.R().SetContext(ctx).SetHeader("Content-Type", "text/plain").Post(url)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send post request")
		return
	}

	log.Info().
		Str("url", url).
		Str("status", resp.Status()).
		Msg("Metric is sent")
}
