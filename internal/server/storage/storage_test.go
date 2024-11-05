package storage_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/db"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

func TestMemStorageInitialization(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())

	assert.NotNil(t, memStorage)
	assert.NotNil(t, memStorage.GetAll(context.Background()))
	assert.Empty(t, memStorage.GetAll(context.Background()))
}

func TestMemStorageUpdateAndGet(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())

	// Prepare metric
	delta := int64(100)
	gaugeValue := float64(20.5)
	metric1 := &db.MtrMetric{
		ID:    domain.MetricName("test_counter"),
		MType: domain.Counter,
		Delta: &delta,
	}
	metric2 := &db.MtrMetric{
		ID:    domain.MetricName("test_gauge"),
		MType: domain.Gauge,
		Value: &gaugeValue,
	}

	// Test updating counter metric
	err := memStorage.Update(context.Background(), metric1)
	require.NoError(t, err)

	// Test updating gauge metric
	err = memStorage.Update(context.Background(), metric2)
	require.NoError(t, err)

	// Check if they are retrievable
	retrievedMetric1, exists := memStorage.Get(context.Background(), domain.MetricName("test_counter"))
	assert.True(t, exists)
	assert.Equal(t, delta, *retrievedMetric1.Delta)

	retrievedMetric2, exists := memStorage.Get(context.Background(), domain.MetricName("test_gauge"))
	assert.True(t, exists)
	assert.InDelta(t, gaugeValue, *retrievedMetric2.Value, 0.0001)
}

func TestMemStorageGetAll(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())

	// Prepare metrics
	delta := int64(150)
	gaugeValue := float64(55.5)
	metric1 := &db.MtrMetric{
		ID:    domain.MetricName("counter_metric"),
		MType: domain.Counter,
		Delta: &delta,
	}
	metric2 := &db.MtrMetric{
		ID:    domain.MetricName("gauge_metric"),
		MType: domain.Gauge,
		Value: &gaugeValue,
	}

	// Update storage
	_ = memStorage.Update(context.Background(), metric1)
	_ = memStorage.Update(context.Background(), metric2)

	// Get all metrics
	allMetrics := memStorage.GetAll(context.Background())

	assert.Len(t, allMetrics, 2)
	assert.Equal(t, delta, *allMetrics["counter_metric"].Delta)
	assert.InDelta(t, gaugeValue, *allMetrics["gauge_metric"].Value, 0000.1)
}

func TestMemStorageBackupAndRestore(t *testing.T) {
	t.Parallel()

	// Setup temporary file for testing backup/restore
	tempDir := t.TempDir()
	// Define the path for the temporary file
	tmpFile := filepath.Join(tempDir, "test_metrics.json")

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
	metric1 := &db.MtrMetric{
		ID:    domain.MetricName("backup_counter"),
		MType: domain.Counter,
		Delta: &delta,
	}
	metric2 := &db.MtrMetric{
		ID:    domain.MetricName("backup_gauge"),
		MType: domain.Gauge,
		Value: &gaugeValue,
	}

	// Update storage and force save
	_ = memStorage.Update(context.Background(), metric1)
	_ = memStorage.Update(context.Background(), metric2)

	// Verify file exists
	_, err := os.Stat(tmpFile)
	require.NoError(t, err)

	// Load from file and verify
	fileContent, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	var restoredData map[domain.MetricName]db.MtrMetric
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
	assert.Len(t, memStorageRestored.GetAll(context.Background()), 2)
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
			metric := &db.MtrMetric{
				ID:    domain.MetricName("concurrent_counter"),
				MType: domain.Counter,
				Delta: &delta,
			}
			_ = memStorage.Update(context.Background(), metric)
		}()
	}

	wg.Wait()
	retrievedMetric, exists := memStorage.Get(context.Background(), "concurrent_counter")
	assert.True(t, exists)
	assert.Equal(t, delta, *retrievedMetric.Delta)
}

func TestMemStorageUpdateWithNoValue(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())
	metric := &db.MtrMetric{
		ID:    domain.MetricName("invalid_metric"),
		MType: domain.Counter,
	}

	err := memStorage.Update(context.Background(), metric)
	require.Error(t, err)
	assert.Equal(t, "no value provided", err.Error())
}

func TestMemStorageCreate(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())

	// Prepare metrics
	delta := int64(150)
	gaugeValue := float64(55.5)
	metric1 := &db.MtrMetric{
		ID:    domain.MetricName("counter_metric"),
		MType: domain.Counter,
		Delta: &delta,
	}
	metric2 := &db.MtrMetric{
		ID:    domain.MetricName("gauge_metric"),
		MType: domain.Gauge,
		Value: &gaugeValue,
	}

	// Update storage
	err := memStorage.Create(context.Background(), metric1)
	require.NoError(t, err)
	err = memStorage.Create(context.Background(), metric2)
	require.NoError(t, err)

	// Get all metrics
	allMetrics := memStorage.GetAll(context.Background())

	assert.Len(t, allMetrics, 2)
	assert.Equal(t, delta, *allMetrics["counter_metric"].Delta)
	assert.InDelta(t, gaugeValue, *allMetrics["gauge_metric"].Value, 0000.1)
}

func TestMemStorageStartBackup(t *testing.T) {
	t.Parallel()

	// Setup temporary file for testing backup/restore
	tempDir := t.TempDir()
	// Define the path for the temporary file
	tmpFile := filepath.Join(tempDir, "test_backup_metrics.json")

	cfg := &config.Config{
		File:             tmpFile,
		StoreInterval:    1, // Set a short interval for testing
		StoreIntervalDur: 1 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	memStorage := storage.NewMemStorage(testutils.GetTLogger()).WithBackup(ctx, cfg)

	// Prepare and update a metric
	delta := int64(200)
	metric := &db.MtrMetric{
		ID:    domain.MetricName("backup_counter_metric"),
		MType: domain.Counter,
		Delta: &delta,
	}

	err := memStorage.Update(context.Background(), metric)
	require.NoError(t, err)

	// Wait a bit to ensure backup happens
	time.Sleep(2 * time.Second)

	cancel()

	// Verify the backup file exists and has the expected metric data
	fileContent, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	var restoredData map[domain.MetricName]db.MtrMetric
	err = json.Unmarshal(fileContent, &restoredData)
	require.NoError(t, err)

	assert.Equal(t, delta, *restoredData["backup_counter_metric"].Delta)
}

func TestMemStorageGetMany(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())

	// Add some metrics
	delta1 := int64(100)
	delta2 := int64(200)
	metric1 := &db.MtrMetric{
		ID:    domain.MetricName("metric1"),
		MType: domain.Counter,
		Delta: &delta1,
	}
	metric2 := &db.MtrMetric{
		ID:    domain.MetricName("metric2"),
		MType: domain.Counter,
		Delta: &delta2,
	}

	_ = memStorage.Update(context.Background(), metric1)
	_ = memStorage.Update(context.Background(), metric2)

	// Test GetMany with multiple metric names
	metricIDs := []domain.MetricName{"metric1", "metric2"}
	retrievedMetrics, err := memStorage.GetMany(context.Background(), metricIDs)
	require.NoError(t, err)

	assert.Len(t, retrievedMetrics, 2)
	assert.Equal(t, delta1, *retrievedMetrics["metric1"].Delta)
	assert.Equal(t, delta2, *retrievedMetrics["metric2"].Delta)
}

func TestMemStorageUpdateMany(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemStorage(testutils.GetTLogger())

	// Prepare multiple metrics
	delta1 := int64(300)
	delta2 := int64(400)
	metric1 := db.MtrMetric{
		ID:    domain.MetricName("update_metric1"),
		MType: domain.Counter,
		Delta: &delta1,
	}
	metric2 := db.MtrMetric{
		ID:    domain.MetricName("update_metric2"),
		MType: domain.Counter,
		Delta: &delta2,
	}

	metricsToUpdate := []db.MtrMetric{metric1, metric2}

	// Use UpdateMany to add both metrics
	err := memStorage.UpdateMany(context.Background(), &metricsToUpdate)
	require.NoError(t, err)

	// Verify both metrics have been added/updated
	ids := []domain.MetricName{"update_metric1", "update_metric2"}
	allMetrics, _ := memStorage.GetMany(context.Background(), ids)
	assert.Len(t, allMetrics, 2)
	assert.Equal(t, delta1, *allMetrics["update_metric1"].Delta)
	assert.Equal(t, delta2, *allMetrics["update_metric2"].Delta)
}

func TestMemStorageConcurrentBackup(t *testing.T) {
	t.Parallel()

	// Setup temporary file for testing backup
	// Setup temporary file for testing backup/restore
	tempDir := t.TempDir()
	// Define the path for the temporary file
	tmpFile := filepath.Join(tempDir, "test_concurrent_backup_metrics.json")

	cfg := &config.Config{
		File:             tmpFile,
		StoreInterval:    1, // Short interval to allow multiple backups
		StoreIntervalDur: 1 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	memStorage := storage.NewMemStorage(testutils.GetTLogger()).WithBackup(ctx, cfg)

	// Prepare metrics
	delta := int64(50)
	for range 5 {
		metric := db.MtrMetric{
			ID:    domain.MetricName("concurrent_backup_metric"),
			MType: domain.Counter,
			Delta: &delta,
		}
		_ = memStorage.Update(context.Background(), &metric)
	}

	// Wait for backups to run
	time.Sleep(3 * time.Second)

	cancel()

	// Check that the backup file exists
	fileContent, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	// Verify the file has the latest data
	var restoredData map[domain.MetricName]db.MtrMetric
	err = json.Unmarshal(fileContent, &restoredData)
	require.NoError(t, err)

	assert.Equal(t, delta, *restoredData["concurrent_backup_metric"].Delta)
}
