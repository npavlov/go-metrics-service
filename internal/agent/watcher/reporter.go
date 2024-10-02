package watcher

import (
	"context"
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/model"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

// Reporter interface defines the contract for sending watcher
type Reporter interface {
	SendMetrics()
	StartReporter(ctx context.Context, cfg *config.Config)
	sendPostRequest(url string)
}

// MetricReporter implements the Reporter interface
type MetricReporter struct {
	metrics *[]model.Metric
	mux     *sync.RWMutex
	addr    string
}

func (mr *MetricReporter) StartReporter(ctx context.Context, cfg *config.Config) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping watcher reporting")
			return
		default:
			// Add your watcher reporting logic here
			mr.SendMetrics()
			time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
		}
	}
}

// NewMetricReporter creates a new instance of MetricReporter
func NewMetricReporter(metrics *[]model.Metric, mux *sync.RWMutex, address string) *MetricReporter {
	return &MetricReporter{
		metrics: metrics,
		mux:     mux,
		addr:    address,
	}
}

// SendMetrics sends the collected watcher to the server
func (mr *MetricReporter) SendMetrics() {
	mr.mux.Lock()
	defer mr.mux.Unlock()

	for _, metric := range *mr.metrics {
		val, found := metric.GetValue()
		if found {
			url := fmt.Sprintf("%s/update/%s/%s/%s", mr.addr, metric.MType, metric.ID, val)
			mr.sendPostRequest(url)
		}
	}
}

// sendPostRequest sends a POST request to the given URL
func (mr *MetricReporter) sendPostRequest(url string) {
	client := resty.New()
	resp, err := client.R().SetHeader("Content-Type", "text/plain").Post(url)
	if err != nil {
		fmt.Println("Error when sending a request:", err)
		return
	}

	fmt.Printf("Metric is sent to %s, status: %s\n", url, resp.Status())
}
