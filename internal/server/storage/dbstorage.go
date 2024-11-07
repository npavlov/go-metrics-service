package storage

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/server/dbmanager"
)

const maxRetries = 3

type DBStorage struct {
	Queries *db.Queries
	l       *zerolog.Logger
	dbCon   dbmanager.PgxPool
}

// NewDBStorage initializes a new DBStorage instance.
func NewDBStorage(dbCon dbmanager.PgxPool, l *zerolog.Logger) *DBStorage {
	return &DBStorage{
		dbCon:   dbCon,
		Queries: db.New(dbCon),
		l:       l,
	}
}

// retryOperation executes a database operation with retry logic.
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

		return backoff.Permanent(err)
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
func (ds *DBStorage) GetAll(ctx context.Context) map[domain.MetricName]db.Metric {
	metrics := make(map[domain.MetricName]db.Metric)

	err := ds.retryOperation(ctx, func() error {
		results, err := ds.Queries.GetAllMetrics(ctx)
		if err != nil {
			ds.l.Error().Err(err).Msg("error getting metrics")

			return errors.Wrap(err, "error getting metrics")
		}

		for _, m := range results {
			metrics[m.ID] = *db.NewMetric(m.ID, m.MType, m.Delta, m.Value)
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
func (ds *DBStorage) Get(ctx context.Context, name domain.MetricName) (*db.Metric, bool) {
	var metric db.Metric

	err := ds.retryOperation(ctx, func() error {
		result, err := ds.Queries.GetUnifiedMetric(ctx, name)
		if err != nil {
			ds.l.Error().Err(err).Msg("failed to retrieve metric")

			return errors.Wrap(err, "failed to retrieve metric")
		}
		metric.FromFields(result.ID, result.MType, result.Delta, result.Value)

		return nil
	})
	if err != nil {
		return nil, false
	}

	return &metric, true
}

// GetMany retrieves multiple metrics based on their names with retry logic.
func (ds *DBStorage) GetMany(ctx context.Context, names []domain.MetricName) (map[domain.MetricName]db.Metric, error) {
	if len(names) == 0 {
		return map[domain.MetricName]db.Metric{}, nil
	}

	nameStrings := make([]string, len(names))
	for i, name := range names {
		nameStrings[i] = string(name)
	}

	metrics := make(map[domain.MetricName]db.Metric)

	err := ds.retryOperation(ctx, func() error {
		results, err := ds.Queries.GetManyMetrics(ctx, nameStrings)
		if err != nil {
			ds.l.Error().Err(err).Msg("error getting multiple metrics")

			return errors.Wrap(err, "error getting multiple metrics")
		}

		for _, m := range results {
			metrics[m.ID] = *db.NewMetric(m.ID, m.MType, m.Delta, m.Value)
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve multiple metrics after retries")
	}

	return metrics, nil
}

// Update modifies an existing metric in the database with retry logic.
func (ds *DBStorage) Update(ctx context.Context, metric *db.Metric) error {
	if metric.Delta == nil && metric.Value == nil {
		return errors.New(errNoValue)
	}

	err := ds.retryOperation(ctx, func() error {
		switch metric.MType {
		case domain.Gauge:
			err := ds.Queries.UpdateGaugeMetric(ctx, db.UpdateGaugeMetricParams{
				Value:    metric.Value,
				MetricID: metric.ID,
			})
			if err != nil {
				ds.l.Error().Err(err).Msg("error updating metric")

				return errors.Wrap(err, "error updating metric")
			}
		case domain.Counter:
			err := ds.Queries.UpdateCounterMetric(ctx, db.UpdateCounterMetricParams{
				Delta:    metric.Delta,
				MetricID: metric.ID,
			})
			if err != nil {
				ds.l.Error().Err(err).Msg("error updating metric")

				return errors.Wrap(err, "error updating metric")
			}
		}

		return nil
	})

	return err
}

// Create inserts a new metric into the database with retry logic.
func (ds *DBStorage) Create(ctx context.Context, metric *db.Metric) error {
	if metric.Delta == nil && metric.Value == nil {
		return errors.New(errNoValue)
	}

	//nolint:exhaustruct
	err := ds.retryOperation(ctx, func() error {
		tx, err := ds.dbCon.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			ds.l.Error().Err(err).Msg("failed to start transaction for Create")

			return errors.Wrap(err, "failed to start transaction for Create")
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback(ctx)
			} else {
				err = tx.Commit(ctx)
			}
		}()

		query := ds.Queries.WithTx(tx)
		err = query.InsertMtrMetric(ctx, db.InsertMtrMetricParams{
			MType: metric.MType,
			ID:    metric.ID,
		})
		if err != nil {
			ds.l.Error().Err(err).Msg("failed to insert metric")

			return errors.Wrap(err, "failed to insert metric")
		}
		switch metric.MType {
		case domain.Gauge:
			err := query.InsertGaugeMetric(ctx, db.InsertGaugeMetricParams{
				Value:    metric.Value,
				MetricID: metric.ID,
			})
			if err != nil {
				ds.l.Error().Err(err).Msg("error insert metric")

				return errors.Wrap(err, "error insert metric")
			}
		case domain.Counter:
			err := query.InsertCounterMetric(ctx, db.InsertCounterMetricParams{
				Delta:    metric.Delta,
				MetricID: metric.ID,
			})
			if err != nil {
				ds.l.Error().Err(err).Msg("error insert metric")

				return errors.Wrap(err, "error insert metric")
			}
		}

		return nil
	})

	return err
}

// UpdateMany updates multiple metrics in the database with retry logic.
func (ds *DBStorage) UpdateMany(ctx context.Context, metrics *[]db.Metric) error {
	//nolint:exhaustruct
	return ds.retryOperation(ctx, func() error {
		tx, err := ds.dbCon.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			ds.l.Error().Err(err).Msg("failed to start transaction for UpdateMany")

			return errors.Wrap(err, "failed to start transaction for UpdateMany")
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback(ctx)
			} else {
				err = tx.Commit(ctx)
			}
		}()

		query := ds.Queries.WithTx(tx)

		for _, metric := range *metrics {
			err := query.UpsertMtrMetric(ctx, db.UpsertMtrMetricParams{
				ID:    metric.ID,
				MType: metric.MType,
			})
			if err != nil {
				ds.l.Error().Err(err).Msg("error in UpsertMetric during UpdateMany")

				return errors.Wrap(err, "error in UpsertMetric during UpdateMany")
			}
			switch metric.MType {
			case domain.Gauge:
				err := query.UpsertGaugeMetric(ctx, db.UpsertGaugeMetricParams{
					Value:    metric.Value,
					MetricID: metric.ID,
				})
				if err != nil {
					ds.l.Error().Err(err).Msg("error insert metric")

					return errors.Wrap(err, "error insert metric")
				}
			case domain.Counter:
				err := query.UpsertCounterMetric(ctx, db.UpsertCounterMetricParams{
					Delta:    metric.Delta,
					MetricID: metric.ID,
				})
				if err != nil {
					ds.l.Error().Err(err).Msg("error insert metric")

					return errors.Wrap(err, "error insert metric")
				}
			}
		}

		return nil
	})
}

// Ping checks the database connection with retry logic.
func (ds *DBStorage) Ping(ctx context.Context) error {
	if ds.dbCon == nil {
		return errors.New("dbCon is nil")
	}

	return errors.Wrap(ds.dbCon.Ping(ctx), "failed to ping db")
}
