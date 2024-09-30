package metrics

import (
	"context"
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/npavlov/go-metrics-service/internal/storage"
)

// Reporter interface defines the contract for sending metrics
type Reporter interface {
	SendMetrics()
	StartReporter(ctx context.Context, cfg *config.Config)
	sendPostRequest(url string)
}

// MetricReporter implements the Reporter interface
type MetricReporter struct {
	Storage storage.Repository
	addr    string
}

func (m *MetricReporter) StartReporter(ctx context.Context, cfg *config.Config) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping metrics reporting")
			return
		default:
			// Add your metrics reporting logic here
			m.SendMetrics()
			time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
		}
	}
}

// NewMetricReporter creates a new instance of MetricReporter
func NewMetricReporter(storage storage.Repository, addr string) *MetricReporter {
	return &MetricReporter{
		Storage: storage,
		addr:    addr,
	}
}

// SendMetrics sends the collected metrics to the server
func (m *MetricReporter) SendMetrics() {
	for name, value := range m.Storage.GetGauges() {
		url := fmt.Sprintf("%s/update/gauge/%s/%g", m.addr, name, value)
		m.sendPostRequest(url)
	}

	for name, value := range m.Storage.GetCounters() {
		url := fmt.Sprintf("%s/update/counter/%s/%d", m.addr, name, value)
		m.sendPostRequest(url)
	}
}

// sendPostRequest sends a POST request to the given URL
func (m *MetricReporter) sendPostRequest(url string) {
	client := resty.New()
	resp, err := client.R().SetHeader("Content-Type", "text/plain").Post(url)
	if err != nil {
		fmt.Println("Error when sending a request:", err)
		return
	}

	fmt.Printf("Metric is sent to %s, status: %s\n", url, resp.Status())
}
