package storage_test

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestMemStorageInitialization(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())

	assert.NotNil(t, memStorage)
	assert.NotNil(t, memStorage.GetAll())
	assert.Empty(t, memStorage.GetAll())
}

func TestMemStorageUpdateAndGet(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())

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
	assert.InDelta(t, gaugeValue, *retrievedMetric2.Value, 0.0001)
}

func TestMemStorageGetAll(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())

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

	assert.Len(t, allMetrics, 2)
	assert.Equal(t, delta, *allMetrics["counter_metric"].Delta)
	assert.InDelta(t, gaugeValue, *allMetrics["gauge_metric"].Value, 0000.1)
}

func TestMemStorageBackupAndRestore(t *testing.T) {
	t.Parallel()

	// Setup temporary file for testing backup/restore
	tmpFile := "test_metrics.json"
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpFile)

	cfg := &config.Config{
		File:           tmpFile,
		StoreInterval:  0,
		RestoreStorage: false,
	}

	// Create metrics to be backed up
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	memStorage := storage.NewMemStorage(testutils.GetTLogger()).WithBackup(ctx, cfg)

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

	assert.Equal(t, delta, *restoredData[("backup_counter")].Delta)
	assert.InDelta(t, gaugeValue, *restoredData[("backup_gauge")].Value, 0.0001)

	cfgRestore := &config.Config{
		File:           tmpFile,
		StoreInterval:  0,
		RestoreStorage: true,
	}
	// Test restore functionality
	memStorageRestored := storage.NewMemStorage(testutils.GetTLogger()).WithBackup(context.Background(), cfgRestore)
	assert.Len(t, memStorageRestored.GetAll(), 2)
}

func TestMemStorageConcurrentUpdate(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())
	delta := int64(1)

	var wg sync.WaitGroup
	for range 100 {
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
	retrievedMetric, exists := memStorage.Get("concurrent_counter")
	assert.True(t, exists)
	assert.Equal(t, delta, *retrievedMetric.Delta)
}

func TestMemStorageUpdateWithNoValue(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())
	metric := &model.Metric{
		ID:    domain.MetricName("invalid_metric"),
		MType: domain.Counter,
	}

	err := memStorage.Update(metric)
	require.Error(t, err)
	assert.Equal(t, "no value provided", err.Error())
}

func TestMemStorageCreate(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())

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
	err := memStorage.Create(metric1)
	require.NoError(t, err)
	err = memStorage.Create(metric2)
	require.NoError(t, err)

	// Get all metrics
	allMetrics := memStorage.GetAll()

	assert.Len(t, allMetrics, 2)
	assert.Equal(t, delta, *allMetrics["counter_metric"].Delta)
	assert.InDelta(t, gaugeValue, *allMetrics["gauge_metric"].Value, 0000.1)
}
