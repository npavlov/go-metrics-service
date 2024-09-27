package metrics

import (
	"errors"
	"github.com/npavlov/go-metrics-service/internal/types"
)

func Validate(stat types.MetricName) error {
	switch stat {
	case types.Alloc, types.BuckHashSys, types.Frees, types.GCCPUFraction, types.GCSys,
		types.HeapAlloc, types.HeapIdle, types.HeapInuse, types.HeapObjects, types.HeapReleased,
		types.HeapSys, types.LastGC, types.Lookups, types.MCacheInuse, types.MCacheSys,
		types.MSpanInuse, types.MSpanSys, types.Mallocs, types.NextGC, types.NumForcedGC,
		types.NumGC, types.OtherSys, types.PauseTotalNs, types.StackInuse, types.StackSys,
		types.Sys, types.TotalAlloc, types.RandomValue, types.PollCount:
		return nil // valid value
	default:
		return errors.New("invalid MetricName value") // invalid value
	}
}
