package storage_test

import (
	"context"
	"encoding/json"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"testing"
)

func TestMemStorageInitialization(t *testing.T) {
	memStorage := storage.NewMemStorage()

	assert.NotNil(t, memStorage)
	assert.NotNil(t, memStorage.GetAll())
	assert.Equal(t, 0, len(memStorage.GetAll()))
}

func TestMemStorageUpdateAndGet(t *testing.T) {
	memStorage := storage.NewMemStorage()

	// Prepare metric
	delta := int64(100)
	gaugeValue := float64(20.5)
	metric1 := &model.Metric{
		ID:    domain.MetricName("test_counter"),
		MType: domain.Counter,
		Delta: &delta,
	}
	metric2 := &model.Metric{
		ID:    domain.MetricName("test_gauge"),
		MType: domain.Gauge,
		Value: &gaugeValue,
	}

	// Test updating counter metric
	err := memStorage.Update(metric1)
	require.NoError(t, err)

	// Test updating gauge metric
	err = memStorage.Update(metric2)
	require.NoError(t, err)

	// Check if they are retrievable
	retrievedMetric1, exists := memStorage.Get(domain.MetricName("test_counter"))
	assert.True(t, exists)
	assert.Equal(t, delta, *retrievedMetric1.Delta)

	retrievedMetric2, exists := memStorage.Get(domain.MetricName("test_gauge"))
	assert.True(t, exists)
	assert.Equal(t, gaugeValue, *retrievedMetric2.Value)
}

func TestMemStorageGetAll(t *testing.T) {
	memStorage := storage.NewMemStorage()

	// Prepare metrics
	delta := int64(150)
	gaugeValue := float64(55.5)
	metric1 := &model.Metric{
		ID:    domain.MetricName("counter_metric"),
		MType: domain.Counter,
		Delta: &delta,
	}
	metric2 := &model.Metric{
		ID:    domain.MetricName("gauge_metric"),
		MType: domain.Gauge,
		Value: &gaugeValue,
	}

	// Update storage
	_ = memStorage.Update(metric1)
	_ = memStorage.Update(metric2)

	// Get all metrics
	allMetrics := memStorage.GetAll()

	assert.Equal(t, 2, len(allMetrics))
	assert.Equal(t, delta, *allMetrics[domain.MetricName("counter_metric")].Delta)
	assert.Equal(t, gaugeValue, *allMetrics[domain.MetricName("gauge_metric")].Value)
}

func TestMemStorageBackupAndRestore(t *testing.T) {
	// Setup temporary file for testing backup/restore
	tmpFile := "test_metrics.json"
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpFile)

	cfg := &config.Config{
		File:          tmpFile,
		StoreInterval: 0,
	}

	// Create metrics to be backed up
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	memStorage := storage.NewMemStorage().WithBackup(ctx, cfg)

	delta := int64(30)
	gaugeValue := float64(42.42)
	metric1 := &model.Metric{
		ID:    domain.MetricName("backup_counter"),
		MType: domain.Counter,
		Delta: &delta,
	}
	metric2 := &model.Metric{
		ID:    domain.MetricName("backup_gauge"),
		MType: domain.Gauge,
		Value: &gaugeValue,
	}

	// Update storage and force save
	_ = memStorage.Update(metric1)
	_ = memStorage.Update(metric2)

	// Verify file exists
	_, err := os.Stat(tmpFile)
	require.NoError(t, err)

	// Load from file and verify
	fileContent, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	var restoredData map[domain.MetricName]model.Metric
	err = json.Unmarshal(fileContent, &restoredData)
	require.NoError(t, err)

	assert.Equal(t, delta, *restoredData[domain.MetricName("backup_counter")].Delta)
	assert.Equal(t, gaugeValue, *restoredData[domain.MetricName("backup_gauge")].Value)

	// Test restore functionality
	memStorageRestored := storage.NewMemStorage().WithBackup(context.Background(), cfg)
	assert.Equal(t, 2, len(memStorageRestored.GetAll()))
}

func TestMemStorageConcurrentUpdate(t *testing.T) {
	memStorage := storage.NewMemStorage()
	delta := int64(1)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metric := &model.Metric{
				ID:    domain.MetricName("concurrent_counter"),
				MType: domain.Counter,
				Delta: &delta,
			}
			_ = memStorage.Update(metric)
		}()
	}

	wg.Wait()
	retrievedMetric, exists := memStorage.Get(domain.MetricName("concurrent_counter"))
	assert.True(t, exists)
	assert.Equal(t, delta, *retrievedMetric.Delta)
}

func TestMemStorageUpdateWithNoValue(t *testing.T) {
	memStorage := storage.NewMemStorage()
	metric := &model.Metric{
		ID:    domain.MetricName("invalid_metric"),
		MType: domain.Counter,
	}

	err := memStorage.Update(metric)
	assert.Error(t, err)
	assert.Equal(t, "no value provided", err.Error())
}
