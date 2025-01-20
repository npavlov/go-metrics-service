package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/logger"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
)

var errPingError = errors.New("ping error")

func setupDBStorage(t *testing.T) (*storage.DBStorage, pgxmock.PgxPoolIface) {
	t.Helper()

	mockDB, err := pgxmock.NewPool()
	require.NoError(t, err)

	log := logger.NewLogger().Get()
	dbStorage := storage.NewDBStorage(mockDB, log)

	return dbStorage, mockDB
}

func TestDBStorage_GetAll(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()

	// Mocking expected rows for the GetAll query
	rows := pgxmock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow(domain.MetricName("metric1"), domain.MetricType("counter"), int64Ptr(10), nil).
		AddRow(domain.MetricName("metric2"), domain.MetricType("gauge"), nil, float64Ptr(3.14))

	mock.ExpectQuery(`SELECT .* FROM mtr_metrics`).WillReturnRows(rows)

	metrics := dbStorage.GetAll(ctx)

	require.NotNil(t, metrics)
	assert.Len(t, metrics, 2)
	assert.Equal(t, int64(10), *metrics["metric1"].Delta)
	assert.InDelta(t, float64(3.14), *metrics["metric2"].Value, 0.0001)
}

func TestDBStorage_Get(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	name := domain.MetricName("metric1")

	// Mocking expected rows for the Get query
	row := pgxmock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow(domain.MetricName("metric1"), domain.MetricType("counter"), int64Ptr(10), nil)

	mock.ExpectQuery("SELECT .* FROM mtr_metrics").
		WithArgs(name).
		WillReturnRows(row)

	metric, found := dbStorage.Get(ctx, name)
	require.True(t, found)
	assert.Equal(t, domain.MetricName("metric1"), metric.ID)
	assert.Equal(t, int64(10), *metric.Delta)
}

func TestDBStorage_GetNotFound(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	name := domain.MetricName("unknown_metric")

	// Expect no rows in result
	mock.ExpectQuery("SELECT .* FROM mtr_metrics").
		WithArgs(name).
		WillReturnError(pgx.ErrNoRows)

	metric, found := dbStorage.Get(ctx, name)
	assert.False(t, found)
	assert.Nil(t, metric)
}

func TestDBStorage_Update(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	metric := db.NewMetric("metric1", domain.Gauge, nil, float64Ptr(200))

	// Mocking an update operation
	mock.ExpectExec("UPDATE gauge_metrics SET").
		WithArgs(metric.ID, metric.Value).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := dbStorage.Update(ctx, metric)
	require.NoError(t, err)

	metric2 := db.NewMetric("metric2", domain.Counter, int64Ptr(100), nil)

	// Mocking an update operation
	mock.ExpectExec("UPDATE counter_metrics SET").
		WithArgs(metric2.ID, metric2.Delta).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = dbStorage.Update(ctx, metric2)
	assert.NoError(t, err)
}

func TestDBStorage_CreateGauge(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	metric := db.NewMetric("metric1", domain.Gauge, nil, float64Ptr(100))

	// Mocking a transaction with upserts
	mock.ExpectBegin()

	mock.ExpectExec("INSERT INTO mtr_metrics").
		WithArgs(metric.ID, metric.MType).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO gauge_metrics").
		WithArgs(metric.ID, metric.Value).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectCommit()

	err := dbStorage.Create(ctx, metric)
	assert.NoError(t, err)
}

func TestDBStorage_CreateCounter(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	metric := db.NewMetric("metric2", domain.Counter, int64Ptr(100), nil)

	// Mocking a transaction with upserts
	mock.ExpectBegin()

	mock.ExpectExec("INSERT INTO mtr_metrics").
		WithArgs(metric.ID, metric.MType).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO counter_metrics").
		WithArgs(metric.ID, metric.Delta).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectCommit()

	err := dbStorage.Create(ctx, metric)
	assert.NoError(t, err)
}

func TestDBStorage_UpdateMany(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	metrics := []db.Metric{
		*db.NewMetric("metric1", domain.Gauge, nil, float64Ptr(10.5)),
		*db.NewMetric("metric2", domain.Gauge, int64Ptr(5), nil),
	}

	// Mocking a transaction with upserts
	mock.ExpectBegin()

	key1, key2 := storage.KeyNameAsHash64("update_many")

	mock.ExpectExec("SELECT pg_advisory_xact_lock").WithArgs(key1, key2).
		WillReturnResult(pgxmock.NewResult("SELECT", 0))

	for _, metric := range metrics {
		mock.ExpectExec("INSERT INTO mtr_metrics").
			WithArgs(metric.ID, metric.MType).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		switch metric.MType {
		case domain.Counter:
			mock.ExpectExec("INSERT INTO counter_metrics").
				WithArgs(metric.ID, metric.Delta).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
		case domain.Gauge:
			mock.ExpectExec("INSERT INTO gauge_metrics").
				WithArgs(metric.ID, metric.Value).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
		}
	}
	mock.ExpectCommit()

	err := dbStorage.UpdateMany(ctx, &metrics)
	assert.NoError(t, err)
}

func TestDBStorage_Ping(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()

	// Mocking a successful ping
	mock.ExpectPing().WillReturnError(nil)

	err := dbStorage.Ping(ctx)
	assert.NoError(t, err)
}

func TestDBStorage_GetMany_Success(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	names := []domain.MetricName{"metric1", "metric2"}

	// Mock expected rows for the successful retrieval
	rows := pgxmock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow(domain.MetricName("metric1"), domain.MetricType("counter"), int64Ptr(10), nil).
		AddRow(domain.MetricName("metric2"), domain.MetricType("gauge"), nil, float64Ptr(3.14))
	mock.ExpectQuery("SELECT .* FROM mtr_metrics").
		WithArgs([]string{"metric1", "metric2"}).
		WillReturnRows(rows)

	metrics, err := dbStorage.GetMany(ctx, names)
	require.NoError(t, err)
	require.Len(t, metrics, 2)
	assert.Equal(t, int64(10), *metrics["metric1"].Delta)
	assert.InDelta(t, float64(3.14), *metrics["metric2"].Value, 0.0001)
}

func TestDBStorage_GetMany_NoMetricsFound(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	names := []domain.MetricName{"unknown_metric1", "unknown_metric2"}

	// Mock expected empty result set for unknown metrics
	mock.ExpectQuery("SELECT .* FROM mtr_metrics").
		WithArgs([]string{"unknown_metric1", "unknown_metric2"}).
		WillReturnRows(pgxmock.NewRows([]string{"id", "type", "delta", "value"}))

	metrics, err := dbStorage.GetMany(ctx, names)
	require.NoError(t, err)
	assert.Empty(t, metrics)
}

func TestDBStorage_GetMany_PartialMetricsFound(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	names := []domain.MetricName{"metric1", "unknown_metric"}

	// Mock expected rows where only one metric is found
	rows := pgxmock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow(domain.MetricName("metric1"), domain.MetricType("counter"), int64Ptr(10), nil)
	mock.ExpectQuery("SELECT .* FROM mtr_metrics").
		WithArgs([]string{"metric1", "unknown_metric"}).
		WillReturnRows(rows)

	metrics, err := dbStorage.GetMany(ctx, names)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	assert.Equal(t, int64(10), *metrics["metric1"].Delta)
}

func TestDBStorage_GetMany_DBError(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	names := []domain.MetricName{"metric_error"}

	// Mock an error from the database
	mock.ExpectQuery("SELECT .* FROM mtr_metrics").
		WithArgs("metric_error").
		WillReturnError(errPingError)

	metrics, err := dbStorage.GetMany(ctx, names)
	require.Error(t, err)
	assert.Nil(t, metrics)
}

func BenchmarkGetAll(b *testing.B) {
	mockDB, err := pgxmock.NewPool()

	if err != nil {
		b.Fatal(err)
	}

	log := logger.NewLogger().Get()
	dbStorage := storage.NewDBStorage(mockDB, log)

	// Mocking expected rows for the GetAll query
	rows := pgxmock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow(domain.MetricName("metric1"), domain.MetricType("counter"), int64Ptr(10), nil).
		AddRow(domain.MetricName("metric2"), domain.MetricType("gauge"), nil, float64Ptr(3.14)).
		AddRow(domain.MetricName("metric3"), domain.MetricType("gauge"), nil, float64Ptr(6.77777)).
		AddRow(domain.MetricName("metric4"), domain.MetricType("counter"), int64Ptr(10), nil).
		AddRow(domain.MetricName("metric5"), domain.MetricType("gauge"), nil, float64Ptr(3.14)).
		AddRow(domain.MetricName("metric6"), domain.MetricType("gauge"), nil, float64Ptr(6.77777)).
		AddRow(domain.MetricName("metric7"), domain.MetricType("counter"), int64Ptr(10), nil).
		AddRow(domain.MetricName("metric8"), domain.MetricType("gauge"), nil, float64Ptr(3.14)).
		AddRow(domain.MetricName("metric9"), domain.MetricType("gauge"), nil, float64Ptr(6.77777)).
		AddRow(domain.MetricName("metric10"), domain.MetricType("counter"), int64Ptr(10), nil).
		AddRow(domain.MetricName("metric11"), domain.MetricType("gauge"), nil, float64Ptr(3.14)).
		AddRow(domain.MetricName("metric12"), domain.MetricType("gauge"), nil, float64Ptr(6.77777))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockDB.ExpectQuery(`SELECT .* FROM mtr_metrics`).WillReturnRows(rows)
		dbStorage.GetAll(context.Background())
	}
}

// Helper function to create float64 pointers.
func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
