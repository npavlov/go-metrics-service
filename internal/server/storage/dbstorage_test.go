package storage_test

import (
	"context"
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

//nolint:ireturn
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

	mock.ExpectQuery("SELECT id, type, delta, value FROM mtr_metrics").WillReturnRows(rows)

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

	mock.ExpectQuery("SELECT id, type, delta, value FROM mtr_metrics WHERE id = \\$1").
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
	mock.ExpectQuery("SELECT id, type, delta, value FROM mtr_metrics WHERE id = \\$1").
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
	metric := &db.MtrMetric{ID: "metric1", MType: domain.Gauge, Value: float64Ptr(200)}

	// Mocking an update operation
	mock.ExpectExec("UPDATE mtr_metrics SET delta = \\$3, value = \\$4 WHERE id = \\$1 AND type = \\$2").
		WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := dbStorage.Update(ctx, metric)
	assert.NoError(t, err)
}

func TestDBStorage_Create(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	metric := &db.MtrMetric{ID: "metric1", MType: domain.Gauge, Value: float64Ptr(100)}

	// Mocking an insert operation
	mock.ExpectExec("INSERT INTO mtr_metrics").
		WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := dbStorage.Create(ctx, metric)
	assert.NoError(t, err)
}

func TestDBStorage_UpdateMany(t *testing.T) {
	t.Parallel()

	dbStorage, mock := setupDBStorage(t)
	defer mock.Close()

	ctx := context.Background()
	metrics := []db.MtrMetric{
		{
			ID:    "metric1",
			MType: domain.Gauge,
			Delta: nil,
			Value: float64Ptr(10.5),
		},
		{
			ID:    "metric2",
			MType: domain.Counter,
			Delta: int64Ptr(5),
			Value: nil,
		},
	}

	// Mocking a transaction with upserts
	mock.ExpectBegin()
	for _, metric := range metrics {
		mock.ExpectExec("INSERT INTO mtr_metrics").
			WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
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

// Helper function to create float64 pointers.
func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
