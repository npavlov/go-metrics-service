package metric_types

import "errors"

type MetricName string

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

const (
	Alloc         MetricName = "Alloc"
	BuckHashSys   MetricName = "BuckHashSys"
	Frees         MetricName = "Frees"
	GCCPUFraction MetricName = "GCCPUFraction"
	GCSys         MetricName = "GCSys"
	HeapAlloc     MetricName = "HeapAlloc"
	HeapIdle      MetricName = "HeapIdle"
	HeapInuse     MetricName = "HeapInuse"
	HeapObjects   MetricName = "HeapObjects"
	HeapReleased  MetricName = "HeapReleased"
	HeapSys       MetricName = "HeapSys"
	LastGC        MetricName = "LastGC"
	Lookups       MetricName = "Lookups"
	MCacheInuse   MetricName = "MCacheInuse"
	MCacheSys     MetricName = "MCacheSys"
	MSpanInuse    MetricName = "MSpanInuse"
	MSpanSys      MetricName = "MSpanSys"
	Mallocs       MetricName = "Mallocs"
	NextGC        MetricName = "NextGC"
	NumForcedGC   MetricName = "NumForcedGC"
	NumGC         MetricName = "NumGC"
	OtherSys      MetricName = "OtherSys"
	PauseTotalNs  MetricName = "PauseTotalNs"
	StackInuse    MetricName = "StackInuse"
	StackSys      MetricName = "StackSys"
	Sys           MetricName = "Sys"
	TotalAlloc    MetricName = "TotalAlloc"
	RandomValue   MetricName = "RandomValue"
	PollCount     MetricName = "PollCount"
)

func Validate(stat MetricName) error {
	switch stat {
	case Alloc, BuckHashSys, Frees, GCCPUFraction, GCSys,
		HeapAlloc, HeapIdle, HeapInuse, HeapObjects, HeapReleased,
		HeapSys, LastGC, Lookups, MCacheInuse, MCacheSys,
		MSpanInuse, MSpanSys, Mallocs, NextGC, NumForcedGC,
		NumGC, OtherSys, PauseTotalNs, StackInuse, StackSys,
		Sys, TotalAlloc, RandomValue, PollCount:
		return nil // valid value
	default:
		return errors.New("invalid MetricName value") // invalid value
	}
}
