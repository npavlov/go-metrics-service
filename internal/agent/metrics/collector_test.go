package metrics

import (
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollector_UpdateMetrics(t *testing.T) {
	st := storage.NewMemStorage()
	var collector = NewMetricCollector(st)

	// Call the method to test
	collector.UpdateMetrics()

	counters := collector.Storage.GetCounters()
	gauges := collector.Storage.GetGauges()

	assert.Equal(t, 1, len(counters))
	assert.Equal(t, 28, len(gauges))

	collector.UpdateMetrics()

	value, ok := collector.Storage.GetCounter(types.PollCount)

	assert.True(t, ok)
	assert.Equal(t, int64(2), value)
}
