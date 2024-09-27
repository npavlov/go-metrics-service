package metrics

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/npavlov/go-metrics-service/internal/agent/metrictypes"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"math/rand"
	"runtime"
)

type Service interface {
	SendMetrics()
	UpdateMetrics()
}

type MetricService struct {
	Storage storage.Repository
	addr    string
}

func NewMetricService(storage storage.Repository, addr string) *MetricService {
	return &MetricService{
		Storage: storage,
		addr:    addr,
	}
}

func (m *MetricService) UpdateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.Storage.UpdateGauge(metrictypes.Alloc, float64(memStats.Alloc))
	m.Storage.UpdateGauge(metrictypes.BuckHashSys, float64(memStats.BuckHashSys))
	m.Storage.UpdateGauge(metrictypes.Frees, float64(memStats.Frees))
	m.Storage.UpdateGauge(metrictypes.GCCPUFraction, memStats.GCCPUFraction)
	m.Storage.UpdateGauge(metrictypes.GCSys, float64(memStats.GCSys))
	m.Storage.UpdateGauge(metrictypes.HeapAlloc, float64(memStats.HeapAlloc))
	m.Storage.UpdateGauge(metrictypes.HeapIdle, float64(memStats.HeapIdle))
	m.Storage.UpdateGauge(metrictypes.HeapInuse, float64(memStats.HeapInuse))
	m.Storage.UpdateGauge(metrictypes.HeapObjects, float64(memStats.HeapObjects))
	m.Storage.UpdateGauge(metrictypes.HeapReleased, float64(memStats.HeapReleased))
	m.Storage.UpdateGauge(metrictypes.HeapSys, float64(memStats.HeapSys))
	m.Storage.UpdateGauge(metrictypes.LastGC, float64(memStats.LastGC))
	m.Storage.UpdateGauge(metrictypes.Lookups, float64(memStats.Lookups))
	m.Storage.UpdateGauge(metrictypes.MCacheInuse, float64(memStats.MCacheInuse))
	m.Storage.UpdateGauge(metrictypes.MCacheSys, float64(memStats.MCacheSys))
	m.Storage.UpdateGauge(metrictypes.MSpanInuse, float64(memStats.MSpanInuse))
	m.Storage.UpdateGauge(metrictypes.MSpanSys, float64(memStats.MSpanSys))
	m.Storage.UpdateGauge(metrictypes.Mallocs, float64(memStats.Mallocs))
	m.Storage.UpdateGauge(metrictypes.NextGC, float64(memStats.NextGC))
	m.Storage.UpdateGauge(metrictypes.NumForcedGC, float64(memStats.NumForcedGC))
	m.Storage.UpdateGauge(metrictypes.NumGC, float64(memStats.NumGC))
	m.Storage.UpdateGauge(metrictypes.PauseTotalNs, float64(memStats.PauseTotalNs))
	m.Storage.UpdateGauge(metrictypes.StackInuse, float64(memStats.StackInuse))
	m.Storage.UpdateGauge(metrictypes.StackSys, float64(memStats.StackSys))
	m.Storage.UpdateGauge(metrictypes.OtherSys, float64(memStats.OtherSys))
	m.Storage.UpdateGauge(metrictypes.Sys, float64(memStats.Sys))
	m.Storage.UpdateGauge(metrictypes.TotalAlloc, float64(memStats.TotalAlloc))
	m.Storage.UpdateGauge(metrictypes.RandomValue, rand.Float64())
	m.Storage.IncCounter(metrictypes.PollCount)
}

func (m *MetricService) SendMetrics() {
	for name, value := range m.Storage.GetGauges() {
		url := fmt.Sprintf("%s/update/gauge/%s/%f", m.addr, name, value)
		m.sendPostRequest(url)
	}

	for name, value := range m.Storage.GetCounters() {
		url := fmt.Sprintf("%s/update/counter/%s/%d", m.addr, name, value)
		m.sendPostRequest(url)
	}
}

// Send Metrics to server
func (m *MetricService) sendPostRequest(url string) {
	client := resty.New()
	resp, err := client.R().SetHeader("Content-Type", "text/plain").Post(url)
	if err != nil {
		fmt.Println("Error when sending a request:", err)
		return
	}

	fmt.Printf("Metric is sent to %s, status: %s\n", url, resp.Status())
}
