package metrics_service

import (
	"fmt"
	. "github.com/npavlov/go-metrics-service/internal/metric-types"
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

	m.storage.UpdateGauge(Alloc, float64(memStats.Alloc))
	m.storage.UpdateGauge(BuckHashSys, float64(memStats.BuckHashSys))
	m.storage.UpdateGauge(Frees, float64(memStats.Frees))
	m.storage.UpdateGauge(GCCPUFraction, memStats.GCCPUFraction)
	m.storage.UpdateGauge(GCSys, float64(memStats.GCSys))
	m.storage.UpdateGauge(HeapAlloc, float64(memStats.HeapAlloc))
	m.storage.UpdateGauge(HeapIdle, float64(memStats.HeapIdle))
	m.storage.UpdateGauge(HeapInuse, float64(memStats.HeapInuse))
	m.storage.UpdateGauge(HeapObjects, float64(memStats.HeapObjects))
	m.storage.UpdateGauge(HeapReleased, float64(memStats.HeapReleased))
	m.storage.UpdateGauge(HeapSys, float64(memStats.HeapSys))
	m.storage.UpdateGauge(LastGC, float64(memStats.LastGC))
	m.storage.UpdateGauge(Lookups, float64(memStats.Lookups))
	m.storage.UpdateGauge(MCacheInuse, float64(memStats.MCacheInuse))
	m.storage.UpdateGauge(MCacheSys, float64(memStats.MCacheSys))
	m.storage.UpdateGauge(MSpanInuse, float64(memStats.MSpanInuse))
	m.storage.UpdateGauge(MSpanSys, float64(memStats.MSpanSys))
	m.storage.UpdateGauge(Mallocs, float64(memStats.Mallocs))
	m.storage.UpdateGauge(NextGC, float64(memStats.NextGC))
	m.storage.UpdateGauge(NumForcedGC, float64(memStats.NumForcedGC))
	m.storage.UpdateGauge(NumGC, float64(memStats.NumGC))
	m.storage.UpdateGauge(PauseTotalNs, float64(memStats.PauseTotalNs))
	m.storage.UpdateGauge(StackInuse, float64(memStats.StackInuse))
	m.storage.UpdateGauge(StackSys, float64(memStats.StackSys))
	m.storage.UpdateGauge(OtherSys, float64(memStats.OtherSys))
	m.storage.UpdateGauge(Sys, float64(memStats.Sys))
	m.storage.UpdateGauge(TotalAlloc, float64(memStats.TotalAlloc))
	m.storage.UpdateGauge(RandomValue, rand.Float64())
	m.storage.IncCounter(PollCount)
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
