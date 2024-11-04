package storage_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/npavlov/go-metrics-service/internal/server/storage"

	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
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
	assert.Equal(suite.T(), int64(10), *metrics["metric1"].Delta)
	assert.Equal(suite.T(), float64(3.14), *metrics["metric2"].Value)

	// Verify all expectations
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *DBStorageTestSuite) TestGet_Success() {
	ctx := context.Background()
	name := domain.MetricName("metric1")

	suite.mock.ExpectQuery(`SELECT id, type, delta, value FROM mtr_metrics WHERE id = \$1`).
		WithArgs(name).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "delta", "value"}).AddRow("metric1", "gauge", nil, 2.71))

	metric, found := suite.storage.Get(ctx, name)

	assert.True(suite.T(), found)
	assert.Equal(suite.T(), domain.MetricName("metric1"), metric.ID)
	assert.Equal(suite.T(), *metric.Value, 2.71)
}

func (suite *DBStorageTestSuite) TestGet_NotFound() {
	ctx := context.Background()
	suite.mock.ExpectQuery(`SELECT id, type, delta, value FROM mtr_metrics WHERE id = \$1`).
		WithArgs("unknown").
		WillReturnError(sql.ErrNoRows)

	_, found := suite.storage.Get(ctx, "unknown")
	suite.False(found)

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// Test UpdateMany method.
func (suite *DBStorageTestSuite) TestUpdateMany_Success() {
	ctx := context.Background()
	metrics := []model.Metric{
		{ID: "metric1", MType: domain.Counter, Delta: int64Ptr(15)},
		{ID: "metric2", MType: domain.Gauge, Value: float64Ptr(42.0)},
	}

	suite.mock.ExpectBegin()
	stmt := suite.mock.ExpectPrepare(`INSERT INTO mtr_metrics`)

	stmt.ExpectExec().
		WithArgs(metrics[0].ID, metrics[0].MType, metrics[0].Delta, metrics[0].Value).
		WillReturnResult(sqlmock.NewResult(1, 1))
	stmt.ExpectExec().
		WithArgs(metrics[1].ID, metrics[1].MType, metrics[1].Delta, metrics[1].Value).
		WillReturnResult(sqlmock.NewResult(1, 1))

	suite.mock.ExpectCommit()

	err := suite.storage.UpdateMany(ctx, &metrics)
	suite.NoError(err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
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

	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), *result["metric1"].Value, 3.14)
}

func (suite *DBStorageTestSuite) TestUpdate_Success() {
	ctx := context.Background()
	metric := &model.Metric{ID: "metric1", MType: domain.Counter, Delta: int64Ptr(10)}

	suite.mock.ExpectExec(`UPDATE mtr_metrics SET delta = \$2, value = \$3 WHERE id = \$1 AND type = \$4`).
		WithArgs(metric.ID, metric.Delta, metric.Value, metric.MType).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := suite.storage.Update(ctx, metric)
	assert.NoError(suite.T(), err)
}

func (suite *DBStorageTestSuite) TestUpdate_RetryOnTransientError() {
	ctx := context.Background()
	metric := &model.Metric{ID: "metric1", MType: domain.Counter, Delta: int64Ptr(10)}

	// First attempt fails
	suite.mock.ExpectExec(`UPDATE mtr_metrics SET delta = \$2, value = \$3 WHERE id = \$1 AND type = \$4`).
		WithArgs(metric.ID, metric.Delta, metric.Value, metric.MType).
		WillReturnError(&pgconn.PgError{
			Code:    pgerrcode.SQLClientUnableToEstablishSQLConnection,
			Message: "could not establish a connection to the database",
		})

	// Second attempt succeeds
	suite.mock.ExpectExec(`UPDATE mtr_metrics SET delta = \$2, value = \$3 WHERE id = \$1 AND type = \$4`).
		WithArgs(metric.ID, metric.Delta, metric.Value, metric.MType).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := suite.storage.Update(ctx, metric)
	assert.NoError(suite.T(), err)
}

func (suite *DBStorageTestSuite) TestCreate_Success() {
	ctx := context.Background()
	metric := &model.Metric{ID: "metric4", MType: domain.Gauge, Value: float64Ptr(2.71)}

	// Flexible regular expression for ExpectExec to handle spaces and special characters
	suite.mock.ExpectExec(`^INSERT INTO mtr_metrics*`).
		WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := suite.storage.Create(ctx, metric)
	assert.NoError(suite.T(), err)

	suite.mock.ExpectQuery(`SELECT id, type, delta, value FROM mtr_metrics WHERE id = \$1`).
		WithArgs(metric.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "delta", "value"}).AddRow("metric1", "gauge", nil, 2.71))

	createdMetric, found := suite.storage.Get(ctx, metric.ID)

	assert.True(suite.T(), found)
	assert.Equal(suite.T(), metric.GetValue(), createdMetric.GetValue())
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
		WithArgs("metric1", "metric2").
		WillReturnRows(rows)

	// Execute the GetMany method
	metrics, err := suite.storage.GetMany(ctx, names)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), metrics, 2)

	// Verify the retrieved metrics
	assert.Equal(suite.T(), int64(10), *metrics["metric1"].Delta)
	assert.Equal(suite.T(), float64(3.14), *metrics["metric2"].Value)
	assert.Nil(suite.T(), metrics["metric1"].Value)
	assert.Nil(suite.T(), metrics["metric2"].Delta)
}

func TestDBStorageTestSuite(t *testing.T) {
	suite.Run(t, new(DBStorageTestSuite))
}

// Utility functions for pointers.
func int64Ptr(i int64) *int64       { return &i }
func float64Ptr(f float64) *float64 { return &f }
