package storage_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/stretchr/testify/suite"

	"github.com/npavlov/go-metrics-service/internal/server/db"

	"github.com/npavlov/go-metrics-service/internal/server/storage"

	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/domain"
)

type DBStorageTestSuite struct {
	suite.Suite
	db      *sql.DB
	mock    sqlmock.Sqlmock
	storage *storage.DBStorage
}

func (suite *DBStorageTestSuite) SetupTest() {
	var err error
	suite.db, suite.mock, err = sqlmock.New()
	suite.Require().NoError(err)

	logger := zerolog.Nop()
	suite.storage = storage.NewDBStorage(suite.db, &logger)
}

func (suite *DBStorageTestSuite) TearDownTest() {
	suite.db.Close()
}

// Test GetAll method.
func (suite *DBStorageTestSuite) TestGetAll_Success() {
	ctx := context.Background()

	// Set up mock database response
	rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow("metric1", "counter", int64(10), nil).
		AddRow("metric2", "gauge", nil, float64(3.14))

	suite.mock.ExpectQuery(`SELECT id, type, delta, value FROM mtr_metrics`).
		WillReturnRows(rows)

	metrics := suite.storage.GetAll(ctx)
	suite.Require().NotNil(metrics)

	// Assertions on returned data
	suite.Equal(int64(10), *metrics["metric1"].Delta)
	suite.InDelta(3.14, *metrics["metric2"].Value, 0.0001)

	// Verify all expectations
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *DBStorageTestSuite) TestGet_Success() {
	ctx := context.Background()
	name := domain.MetricName("metric1")

	suite.mock.ExpectQuery(`SELECT id, type, delta, value FROM mtr_metrics WHERE id = \$1`).
		WithArgs(name).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "delta", "value"}).AddRow("metric1", "gauge", nil, 2.71))

	metric, found := suite.storage.Get(ctx, name)

	suite.True(found)
	suite.Equal(domain.MetricName("metric1"), metric.ID)
	suite.InDelta(2.71, *metric.Value, 0.0001)
}

func (suite *DBStorageTestSuite) TestGet_NotFound() {
	ctx := context.Background()
	suite.mock.ExpectQuery(`SELECT id, type, delta, value FROM mtr_metrics WHERE id = \$1`).
		WithArgs("unknown").
		WillReturnError(sql.ErrNoRows)

	_, found := suite.storage.Get(ctx, "unknown")
	suite.False(found)

	suite.NoError(suite.mock.ExpectationsWereMet())
}

// Test UpdateMany method.
func (suite *DBStorageTestSuite) TestUpdateMany_Success() {
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

	suite.mock.ExpectBegin()
	// Setting up the expected query for each metric update
	for _, metric := range metrics {
		suite.mock.ExpectExec(`INSERT INTO mtr_metrics`).
			WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).
			WillReturnResult(sqlmock.NewResult(0, 1)) // Simulate successful update
	}
	suite.mock.ExpectCommit()

	err := suite.storage.UpdateMany(context.Background(), &metrics)
	suite.NoError(err)
}

func (suite *DBStorageTestSuite) TestGetAll_RetryOnTransientError() {
	ctx := context.Background()

	// First attempt should fail to trigger retry
	suite.mock.ExpectQuery(`SELECT id, type, delta, value FROM mtr_metrics`).
		WillReturnError(&pgconn.PgError{
			Code:    pgerrcode.SQLClientUnableToEstablishSQLConnection,
			Message: "could not establish a connection to the database",
		})

	// Second attempt should succeed
	rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow("metric1", "gauge", nil, 3.14)
	suite.mock.ExpectQuery(`SELECT id, type, delta, value FROM mtr_metrics`).
		WillReturnRows(rows)

	result := suite.storage.GetAll(ctx)

	suite.Len(result, 1)
	suite.InDelta(3.14, *result["metric1"].Value, 0.0001)
}

func (suite *DBStorageTestSuite) TestUpdate_Success() {
	ctx := context.Background()
	metric := &db.MtrMetric{ID: "metric1", MType: domain.Counter, Delta: int64Ptr(10)}

	suite.mock.ExpectExec(`UPDATE mtr_metrics SET`).
		WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := suite.storage.Update(ctx, metric)
	suite.NoError(err)
}

func (suite *DBStorageTestSuite) TestUpdate_RetryOnTransientError() {
	ctx := context.Background()
	metric := &db.MtrMetric{ID: "metric1", MType: domain.Counter, Delta: int64Ptr(10)}

	// First attempt fails
	suite.mock.ExpectExec(`UPDATE mtr_metrics SET`).
		WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).
		WillReturnError(&pgconn.PgError{
			Code:    pgerrcode.SQLClientUnableToEstablishSQLConnection,
			Message: "could not establish a connection to the database",
		})

	// Second attempt succeeds
	suite.mock.ExpectExec(`UPDATE mtr_metrics SET`).
		WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := suite.storage.Update(ctx, metric)
	suite.NoError(err)
}

func (suite *DBStorageTestSuite) TestCreate_Success() {
	ctx := context.Background()
	metric := &db.MtrMetric{ID: "metric4", MType: domain.Gauge, Value: float64Ptr(2.71)}

	// Flexible regular expression for ExpectExec to handle spaces and special characters
	suite.mock.ExpectExec(`INSERT INTO mtr_metrics`).
		WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := suite.storage.Create(ctx, metric)
	suite.Require().NoError(err)

	suite.mock.ExpectQuery(`SELECT id, type, delta, value FROM mtr_metrics WHERE id = \$1`).
		WithArgs(metric.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "delta", "value"}).AddRow("metric1", "gauge", nil, 2.71))

	createdMetric, found := suite.storage.Get(ctx, metric.ID)

	suite.True(found)
	suite.Equal(metric.GetValue(), createdMetric.GetValue())
}

func (suite *DBStorageTestSuite) TestGetMany_Success() {
	ctx := context.Background()
	names := []domain.MetricName{"metric1", "metric2"}

	// Set up expected rows to return
	rows := suite.mock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow("metric1", "counter", int64(10), nil).
		AddRow("metric2", "gauge", nil, float64(3.14))

	// Define the expected SQL query and arguments
	suite.mock.ExpectQuery(`SELECT id, type, delta, value FROM mtr_metrics`).
		WithArgs(pq.Array([]string{"metric1", "metric2"})).
		WillReturnRows(rows)

	// Execute the GetMany method
	metrics, err := suite.storage.GetMany(ctx, names)
	suite.Require().NoError(err)
	suite.Len(metrics, 2)

	// Verify the retrieved metrics
	suite.Equal(int64(10), *metrics["metric1"].Delta)
	suite.InDelta(float64(3.14), *metrics["metric2"].Value, 0.0001)
	suite.Nil(metrics["metric1"].Value)
	suite.Nil(metrics["metric2"].Delta)
}

func TestDBStorageTestSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(DBStorageTestSuite))
}

// Utility functions for pointers.
func int64Ptr(i int64) *int64       { return &i }
func float64Ptr(f float64) *float64 { return &f }
