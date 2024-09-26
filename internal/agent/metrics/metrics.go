package metrics

import (
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/agent/metrictypes"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"math/rand"
	"net/http"
	"runtime"
)

const addr = "http://localhost:8080"

type Service interface {
	SendMetrics()
	UpdateMetrics()
}

type MetricService struct {
	storage storage.Repository
}

func NewMetricService() *MetricService {
	return &MetricService{
		storage: storage.NewMemStorage(),
	}
}

func (m MetricService) SendMetrics() {
	for name, value := range m.storage.GetGauges() {
		url := fmt.Sprintf("%s/update/gauge/%s/%f", addr, name, value)
		sendPostRequest(url)
	}

	for name, value := range m.storage.GetCounters() {
		url := fmt.Sprintf("%s/update/counter/%s/%d", addr, name, value)
		sendPostRequest(url)
	}
}

func (m MetricService) UpdateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.storage.UpdateGauge(metrictypes.Alloc, float64(memStats.Alloc))
	m.storage.UpdateGauge(metrictypes.BuckHashSys, float64(memStats.BuckHashSys))
	m.storage.UpdateGauge(metrictypes.Frees, float64(memStats.Frees))
	m.storage.UpdateGauge(metrictypes.GCCPUFraction, memStats.GCCPUFraction)
	m.storage.UpdateGauge(metrictypes.GCSys, float64(memStats.GCSys))
	m.storage.UpdateGauge(metrictypes.HeapAlloc, float64(memStats.HeapAlloc))
	m.storage.UpdateGauge(metrictypes.HeapIdle, float64(memStats.HeapIdle))
	m.storage.UpdateGauge(metrictypes.HeapInuse, float64(memStats.HeapInuse))
	m.storage.UpdateGauge(metrictypes.HeapObjects, float64(memStats.HeapObjects))
	m.storage.UpdateGauge(metrictypes.HeapReleased, float64(memStats.HeapReleased))
	m.storage.UpdateGauge(metrictypes.HeapSys, float64(memStats.HeapSys))
	m.storage.UpdateGauge(metrictypes.LastGC, float64(memStats.LastGC))
	m.storage.UpdateGauge(metrictypes.Lookups, float64(memStats.Lookups))
	m.storage.UpdateGauge(metrictypes.MCacheInuse, float64(memStats.MCacheInuse))
	m.storage.UpdateGauge(metrictypes.MCacheSys, float64(memStats.MCacheSys))
	m.storage.UpdateGauge(metrictypes.MSpanInuse, float64(memStats.MSpanInuse))
	m.storage.UpdateGauge(metrictypes.MSpanSys, float64(memStats.MSpanSys))
	m.storage.UpdateGauge(metrictypes.Mallocs, float64(memStats.Mallocs))
	m.storage.UpdateGauge(metrictypes.NextGC, float64(memStats.NextGC))
	m.storage.UpdateGauge(metrictypes.NumForcedGC, float64(memStats.NumForcedGC))
	m.storage.UpdateGauge(metrictypes.NumGC, float64(memStats.NumGC))
	m.storage.UpdateGauge(metrictypes.PauseTotalNs, float64(memStats.PauseTotalNs))
	m.storage.UpdateGauge(metrictypes.StackInuse, float64(memStats.StackInuse))
	m.storage.UpdateGauge(metrictypes.StackSys, float64(memStats.StackSys))
	m.storage.UpdateGauge(metrictypes.OtherSys, float64(memStats.OtherSys))
	m.storage.UpdateGauge(metrictypes.Sys, float64(memStats.Sys))
	m.storage.UpdateGauge(metrictypes.TotalAlloc, float64(memStats.TotalAlloc))
	m.storage.UpdateGauge(metrictypes.RandomValue, rand.Float64())
	m.storage.IncCounter(metrictypes.PollCount)
}

// Функция отправки POST-запроса
func sendPostRequest(url string) {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при отправке запроса:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Метрика отправлена на %s, статус: %s\n", url, resp.Status)
}
