package metrics

import (
	"errors"
	"github.com/npavlov/go-metrics-service/internal/agent/metrictypes"
)

func Validate(stat metrictypes.MetricName) error {
	switch stat {
	case metrictypes.Alloc, metrictypes.BuckHashSys, metrictypes.Frees, metrictypes.GCCPUFraction, metrictypes.GCSys,
		metrictypes.HeapAlloc, metrictypes.HeapIdle, metrictypes.HeapInuse, metrictypes.HeapObjects, metrictypes.HeapReleased,
		metrictypes.HeapSys, metrictypes.LastGC, metrictypes.Lookups, metrictypes.MCacheInuse, metrictypes.MCacheSys,
		metrictypes.MSpanInuse, metrictypes.MSpanSys, metrictypes.Mallocs, metrictypes.NextGC, metrictypes.NumForcedGC,
		metrictypes.NumGC, metrictypes.OtherSys, metrictypes.PauseTotalNs, metrictypes.StackInuse, metrictypes.StackSys,
		metrictypes.Sys, metrictypes.TotalAlloc, metrictypes.RandomValue, metrictypes.PollCount:
		return nil // valid value
	default:
		return errors.New("invalid MetricName value") // invalid value
	}
}
