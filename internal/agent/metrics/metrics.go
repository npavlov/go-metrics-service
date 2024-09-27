package metrics

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
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

	m.Storage.UpdateGauge(types.Alloc, float64(memStats.Alloc))
	m.Storage.UpdateGauge(types.BuckHashSys, float64(memStats.BuckHashSys))
	m.Storage.UpdateGauge(types.Frees, float64(memStats.Frees))
	m.Storage.UpdateGauge(types.GCCPUFraction, memStats.GCCPUFraction)
	m.Storage.UpdateGauge(types.GCSys, float64(memStats.GCSys))
	m.Storage.UpdateGauge(types.HeapAlloc, float64(memStats.HeapAlloc))
	m.Storage.UpdateGauge(types.HeapIdle, float64(memStats.HeapIdle))
	m.Storage.UpdateGauge(types.HeapInuse, float64(memStats.HeapInuse))
	m.Storage.UpdateGauge(types.HeapObjects, float64(memStats.HeapObjects))
	m.Storage.UpdateGauge(types.HeapReleased, float64(memStats.HeapReleased))
	m.Storage.UpdateGauge(types.HeapSys, float64(memStats.HeapSys))
	m.Storage.UpdateGauge(types.LastGC, float64(memStats.LastGC))
	m.Storage.UpdateGauge(types.Lookups, float64(memStats.Lookups))
	m.Storage.UpdateGauge(types.MCacheInuse, float64(memStats.MCacheInuse))
	m.Storage.UpdateGauge(types.MCacheSys, float64(memStats.MCacheSys))
	m.Storage.UpdateGauge(types.MSpanInuse, float64(memStats.MSpanInuse))
	m.Storage.UpdateGauge(types.MSpanSys, float64(memStats.MSpanSys))
	m.Storage.UpdateGauge(types.Mallocs, float64(memStats.Mallocs))
	m.Storage.UpdateGauge(types.NextGC, float64(memStats.NextGC))
	m.Storage.UpdateGauge(types.NumForcedGC, float64(memStats.NumForcedGC))
	m.Storage.UpdateGauge(types.NumGC, float64(memStats.NumGC))
	m.Storage.UpdateGauge(types.PauseTotalNs, float64(memStats.PauseTotalNs))
	m.Storage.UpdateGauge(types.StackInuse, float64(memStats.StackInuse))
	m.Storage.UpdateGauge(types.StackSys, float64(memStats.StackSys))
	m.Storage.UpdateGauge(types.OtherSys, float64(memStats.OtherSys))
	m.Storage.UpdateGauge(types.Sys, float64(memStats.Sys))
	m.Storage.UpdateGauge(types.TotalAlloc, float64(memStats.TotalAlloc))
	m.Storage.UpdateGauge(types.RandomValue, rand.Float64())
	m.Storage.IncCounter(types.PollCount)
}

func (m *MetricService) SendMetrics() {
	for name, value := range m.Storage.GetGauges() {
		url := fmt.Sprintf("%s/update/gauge/%s/%g", m.addr, name, value)
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
