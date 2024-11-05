package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

const (
	maxRetries = 3
)

type DBStorage struct {
	Queries *db.Queries
	l       *zerolog.Logger
	dbCon   *sql.DB
}

// NewDBStorage initializes a new DBStorage instance.
func NewDBStorage(dbCon *sql.DB, l *zerolog.Logger) *DBStorage {
	return &DBStorage{
		dbCon:   dbCon,
		Queries: db.New(dbCon),
		l:       l,
	}
}

// retryOperation executes a database operation with retry logic, handling transient errors automatically.
func (ds *DBStorage) retryOperation(ctx context.Context, operation func() error) error {
	backoffConfig := backoff.NewExponentialBackOff()
	backoffConfig.InitialInterval = 1 * time.Second
	backoffConfig.Multiplier = 3
	retryWithLimit := backoff.WithMaxRetries(backoffConfig, maxRetries)

	err := backoff.Retry(func() error {
		err := operation()
		if err != nil && ds.isRetryableError(err) {
			ds.l.Warn().Err(err).Msg("transient error, retrying operation")

			return err
		}

		return backoff.Permanent(err) // Stop retrying on non-retryable error
	}, backoff.WithContext(retryWithLimit, ctx))

	return errors.Wrap(err, "failed to execute operation after retry")
}

// isRetryableError checks if an error is retryable based on PostgreSQL-specific error codes.
func (ds *DBStorage) isRetryableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.CannotConnectNow,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown:
			return true
		}
	}

	return false
}

// GetAll retrieves all metrics from the database with retry logic.
func (ds *DBStorage) GetAll(ctx context.Context) map[domain.MetricName]db.MtrMetric {
	metrics := make(map[domain.MetricName]db.MtrMetric)

	err := ds.retryOperation(ctx, func() error {
		results, err := ds.Queries.GetAllMetrics(ctx)
		if err != nil {
			ds.l.Error().Err(err).Msg("error getting metrics")

			return errors.Wrap(err, "error getting metrics")
		}

		for _, m := range results {
			metrics[m.ID] = m
		}

		return nil
	})
	if err != nil {
		ds.l.Error().Err(err).Msg("failed to retrieve metrics after retries")

		return nil
	}

	return metrics
}

// Get retrieves a single metric by its name with retry logic.
func (ds *DBStorage) Get(ctx context.Context, name domain.MetricName) (*db.MtrMetric, bool) {
	var metric db.MtrMetric

	err := ds.retryOperation(ctx, func() error {
		result, err := ds.Queries.GetMetric(ctx, name)
		if err != nil {
			ds.l.Error().Err(err).Msg("failed to retrieve metric")

			return errors.Wrap(err, "failed to retrieve metric")
		}
		metric = result

		return nil
	})
	if err != nil {
		return nil, false
	}

	return &metric, true
}

// GetMany retrieves multiple metrics based on their names with retry logic.
//
//nolint:lll
func (ds *DBStorage) GetMany(ctx context.Context, names []domain.MetricName) (map[domain.MetricName]db.MtrMetric, error) {
	if len(names) == 0 {
		return map[domain.MetricName]db.MtrMetric{}, nil
	}

	nameStrings := make([]string, len(names))
	for i, name := range names {
		nameStrings[i] = string(name)
	}

	metrics := make(map[domain.MetricName]db.MtrMetric)

	err := ds.retryOperation(ctx, func() error {
		results, err := ds.Queries.GetManyMetrics(ctx, nameStrings)
		if err != nil {
			ds.l.Error().Err(err).Msg("error getting multiple metrics")

			return errors.Wrap(err, "error getting multiple metrics")
		}

		for _, m := range results {
			metrics[m.ID] = m
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve multiple metrics after retries")
	}

	return metrics, nil
}

// Update modifies an existing metric in the database with retry logic.
func (ds *DBStorage) Update(ctx context.Context, metric *db.MtrMetric) error {
	return ds.retryOperation(ctx, func() error {
		err := ds.Queries.UpdateMetric(ctx, metric.ToUpdateMetricParams())
		if err != nil {
			ds.l.Error().Err(err).Msg("error updating metric")
		}

		return errors.Wrap(err, "error updating metric")
	})
}

// Create inserts a new metric into the database with retry logic.
func (ds *DBStorage) Create(ctx context.Context, metric *db.MtrMetric) error {
	return ds.retryOperation(ctx, func() error {
		err := ds.Queries.InsertMetric(ctx, metric.ToInsertMetricParams())
		if err != nil {
			ds.l.Error().Err(err).Msg("error inserting metric")
		}

		return errors.Wrap(err, "error inserting metric")
	})
}

// UpdateMany updates multiple metrics in the database with retry logic.
func (ds *DBStorage) UpdateMany(ctx context.Context, metrics *[]db.MtrMetric) error {
	return ds.retryOperation(ctx, func() error {
		tx, err := ds.dbCon.BeginTx(ctx, nil)
		if err != nil {
			ds.l.Error().Err(err).Msg("failed to start transaction for UpdateMany")

			return errors.Wrap(err, "failed to start transaction for UpdateMany")
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}()

		q := ds.Queries.WithTx(tx)

		for _, metric := range *metrics {
			err := q.UpsertMetric(ctx, metric.ToUpsertMetricParams())
			if err != nil {
				ds.l.Error().Err(err).Msg("error in UpsertMetric during UpdateMany")

				return errors.Wrap(err, "error in UpsertMetric during UpdateMany")
			}
		}

		return nil
	})
}

// Ping checks the database connection with retry logic.
func (ds *DBStorage) Ping() error {
	if ds.dbCon == nil {
		return errors.New("dbCon is nil")
	}

	return errors.Wrap(ds.dbCon.Ping(), "failed to ping db")
}
