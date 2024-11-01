package storage_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	err = db.AutoMigrate(&model.Metric{})
	require.NoError(t, err, "failed to migrate test database schema")

	return db
}

func TestDBStorageCreate(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := storage.NewDBStorage(db)

	// Prepare test metric
	metric := &model.Metric{
		ID:    domain.MetricName("test_create_metric"),
		MType: domain.Counter,
		Delta: new(int64),
	}
	*metric.Delta = 10

	// Test creating metric
	err := repo.Create(context.Background(), metric)
	require.NoError(t, err, "failed to create metric")

	// Verify metric in DB
	var retrievedMetric model.Metric
	err = db.First(&retrievedMetric, "id = ?", metric.ID).Error
	require.NoError(t, err, "failed to retrieve metric")
	assert.Equal(t, metric.ID, retrievedMetric.ID)
	assert.Equal(t, *metric.Delta, *retrievedMetric.Delta)
}

func TestDBStorageGet(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := storage.NewDBStorage(db)

	// Seed a metric in the database
	metric := model.Metric{
		ID:    domain.MetricName("test_get_metric"),
		MType: domain.Counter,
		Delta: new(int64),
	}
	*metric.Delta = 20
	require.NoError(t, db.Create(&metric).Error, "failed to seed metric")

	// Test retrieving the metric
	retrievedMetric, exists := repo.Get(context.Background(), metric.ID)
	assert.True(t, exists)
	assert.Equal(t, metric.ID, retrievedMetric.ID)
	assert.Equal(t, *metric.Delta, *retrievedMetric.Delta)
}

func TestDBStorageGetMany(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := storage.NewDBStorage(db)

	// Seed multiple metrics
	metrics := []model.Metric{
		{ID: domain.MetricName("metric1"), MType: domain.Counter, Delta: new(int64)},
		{ID: domain.MetricName("metric2"), MType: domain.Gauge, Value: new(float64)},
	}
	*metrics[0].Delta = 100
	*metrics[1].Value = 50.5
	require.NoError(t, db.Create(&metrics[0]).Error, "failed to seed metric1")
	require.NoError(t, db.Create(&metrics[1]).Error, "failed to seed metric2")

	// Test GetMany
	names := []domain.MetricName{"metric1", "metric2"}
	retrievedMetrics, err := repo.GetMany(context.Background(), names)
	require.NoError(t, err)
	assert.Len(t, retrievedMetrics, 2)
	assert.Equal(t, *metrics[0].Delta, *retrievedMetrics["metric1"].Delta)
	assert.InDelta(t, *metrics[1].Value, *retrievedMetrics["metric2"].Value, 0.0001)
}

func TestDBStorageUpdate(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := storage.NewDBStorage(db)

	// Seed a metric
	metric := model.Metric{
		ID:    domain.MetricName("test_update_metric"),
		MType: domain.Counter,
		Delta: new(int64),
	}
	*metric.Delta = 30
	require.NoError(t, db.Create(&metric).Error, "failed to seed metric")

	// Update the metric
	*metric.Delta = 40
	err := repo.Update(context.Background(), &metric)
	require.NoError(t, err, "failed to update metric")

	// Verify update in DB
	var updatedMetric model.Metric
	err = db.First(&updatedMetric, "id = ?", metric.ID).Error
	require.NoError(t, err, "failed to retrieve updated metric")
	assert.Equal(t, *metric.Delta, *updatedMetric.Delta)
}

func TestDBStorageUpdateMany(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := storage.NewDBStorage(db)

	// Seed initial metrics
	metrics := []model.Metric{
		{ID: domain.MetricName("metric_bulk1"), MType: domain.Counter, Delta: new(int64)},
		{ID: domain.MetricName("metric_bulk2"), MType: domain.Gauge, Value: new(float64)},
	}
	*metrics[0].Delta = 300
	*metrics[1].Value = 150.5
	require.NoError(t, db.Create(&metrics[0]).Error, "failed to seed metric_bulk1")
	require.NoError(t, db.Create(&metrics[1]).Error, "failed to seed metric_bulk2")

	// Update metrics in bulk
	*metrics[0].Delta = 400
	*metrics[1].Value = 250.5
	err := repo.UpdateMany(context.Background(), &metrics)
	require.NoError(t, err, "failed to bulk update metrics")

	// Verify updates in DB
	var updatedMetrics []model.Metric
	err = db.Find(&updatedMetrics, "id IN ?", []string{"metric_bulk1", "metric_bulk2"}).Error
	require.NoError(t, err, "failed to retrieve bulk-updated metrics")
	assert.Len(t, updatedMetrics, 2)
	assert.Equal(t, *metrics[0].Delta, *updatedMetrics[0].Delta)
	assert.InDelta(t, *metrics[1].Value, *updatedMetrics[1].Value, 0.0001)
}

func TestDBStoragePing(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := storage.NewDBStorage(db)

	// Test Ping with a valid connection
	err := repo.Ping()
	require.NoError(t, err, "failed to ping database")

	// Test Ping with a nil connection
	repoNil := storage.NewDBStorage(nil)
	err = repoNil.Ping()
	assert.Error(t, err, "expected error on ping with nil DB connection")
}
