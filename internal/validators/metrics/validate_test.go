package metrics

import (
	"github.com/npavlov/go-metrics-service/internal/agent/metrics"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricService_UpdateMetrics(t *testing.T) {
	var memStorage storage.Repository = storage.NewMemStorage()
	metricService := metrics.NewMetricCollector(memStorage)

	// Call the method to test
	metricService.UpdateMetrics()

	counters := metricService.Storage.GetCounters()
	gauges := metricService.Storage.GetGauges()

	for k := range counters {
		err := Validate(k)
		assert.Nil(t, err)
	}

	for k := range gauges {
		err := Validate(k)
		assert.Nil(t, err)
	}

	// Has error if unknown key
	err := Validate("blablabla")

	assert.NotNil(t, err)

}
