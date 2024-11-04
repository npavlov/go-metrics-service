package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
)

type DBStorage struct {
	Db *sql.DB
	l  *zerolog.Logger
}

// NewDBStorage initializes a new DBStorage instance.
func NewDBStorage(db *sql.DB, l *zerolog.Logger) *DBStorage {
	return &DBStorage{
		Db: db,
		l:  l,
	}
}

// retryOperation is a helper function to execute a database operation with retry logic.
func (ds *DBStorage) retryOperation(ctx context.Context, operation func() error) error {
	// Configure exponential backoff strategy
	backoffConfig := backoff.NewExponentialBackOff()
	backoffConfig.InitialInterval = 1 * time.Second
	backoffConfig.Multiplier = 3 // To approximate delays of 1s, 3s, and 5s

	// Limit to a maximum of 3 retries
	retryWithLimit := backoff.WithMaxRetries(backoffConfig, 3)

	// Run the operation with retry logic
	return backoff.Retry(operation, backoff.WithContext(retryWithLimit, ctx))
}

// isRetriableError determines if an error is retriable by checking PostgreSQL-specific error codes.
func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Check for retriable PostgreSQL error codes, such as connection errors.
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
func (ds *DBStorage) GetAll(ctx context.Context) map[domain.MetricName]model.Metric {
	query := `SELECT id, type, delta, value FROM mtr_metrics`
	metrics := make(map[domain.MetricName]model.Metric)

	err := ds.retryOperation(ctx, func() error {
		rows, err := ds.Db.QueryContext(ctx, query)
		if err != nil {
			if isRetriableError(err) {
				ds.l.Warn().Err(err).Msg("transient error on QueryContext, retrying")
				return err
			}
			ds.l.Error().Err(err).Msg("error getting metrics")
			return backoff.Permanent(err)
		}
		defer rows.Close()

		for rows.Next() {
			var metric model.Metric
			if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
				ds.l.Error().Err(err).Msg("error scanning metric")
				return backoff.Permanent(err)
			}
			metrics[metric.ID] = metric
		}

		return rows.Err()
	})
	if err != nil {
		ds.l.Error().Err(err).Msg("failed to retrieve metrics after retries")
		return nil
	}

	return metrics
}

// Get retrieves a single metric by its name with retry logic.
func (ds *DBStorage) Get(ctx context.Context, name domain.MetricName) (*model.Metric, bool) {
	query := `SELECT id, type, delta, value FROM mtr_metrics WHERE id = $1`
	var metric model.Metric

	err := ds.retryOperation(ctx, func() error {
		row := ds.Db.QueryRowContext(ctx, query, name)
		err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			if isRetriableError(err) {
				ds.l.Warn().Err(err).Msg("transient error on QueryRowContext, retrying")
				return err
			}
			ds.l.Error().Err(err).Msg("failed to retrieve metric")
			return backoff.Permanent(err)
		}
		return nil
	})
	if err != nil {
		return nil, false
	}

	return &metric, true
}

// GetMany retrieves multiple metrics based on their names with retry logic.
func (ds *DBStorage) GetMany(ctx context.Context, names []domain.MetricName) (map[domain.MetricName]model.Metric, error) {
	if len(names) == 0 {
		return map[domain.MetricName]model.Metric{}, nil
	}

	placeholders := make([]string, len(names))
	args := make([]interface{}, len(names))
	for i, name := range names {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = name
	}

	query := fmt.Sprintf(`SELECT id, type, delta, value FROM mtr_metrics WHERE id IN (%s)`, strings.Join(placeholders, ", "))
	metrics := make(map[domain.MetricName]model.Metric)

	err := ds.retryOperation(ctx, func() error {
		rows, err := ds.Db.QueryContext(ctx, query, args...)
		if err != nil {
			if isRetriableError(err) {
				ds.l.Warn().Err(err).Msg("transient error on QueryContext, retrying")
				return err
			}
			return backoff.Permanent(err)
		}
		defer rows.Close()

		for rows.Next() {
			var metric model.Metric
			if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
				return backoff.Permanent(err)
			}
			metrics[metric.ID] = metric
		}
		return rows.Err()
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve multiple metrics after retries")
	}

	return metrics, nil
}

// Update modifies an existing metric in the database with retry logic.
func (ds *DBStorage) Update(ctx context.Context, metric *model.Metric) error {
	query := `
		UPDATE mtr_metrics 
		SET delta = $2, value = $3 
		WHERE id = $1 AND type = $4`

	return ds.retryOperation(ctx, func() error {
		_, err := ds.Db.ExecContext(ctx, query, metric.ID, metric.Delta, metric.Value, metric.MType)
		if err != nil {
			if isRetriableError(err) {
				ds.l.Warn().Err(err).Msg("transient error on ExecContext, retrying")
				return err
			}
			return backoff.Permanent(err)
		}
		return nil
	})
}

// Create inserts a new metric into the database with retry logic.
func (ds *DBStorage) Create(ctx context.Context, metric *model.Metric) error {
	query := `
		INSERT INTO mtr_metrics (id, type, delta, value) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id, type) DO NOTHING`

	return ds.retryOperation(ctx, func() error {
		_, err := ds.Db.ExecContext(ctx, query, metric.ID, metric.MType, metric.Delta, metric.Value)
		if err != nil {
			if isRetriableError(err) {
				ds.l.Warn().Err(err).Msg("transient error on ExecContext, retrying")
				return err
			}
			return backoff.Permanent(err)
		}
		return nil
	})
}

// UpdateMany updates multiple metrics in the database with retry logic.
func (ds *DBStorage) UpdateMany(ctx context.Context, metrics *[]model.Metric) error {
	return ds.retryOperation(ctx, func() error {
		tx, err := ds.Db.BeginTx(ctx, nil)
		if err != nil {
			return backoff.Permanent(errors.Wrap(err, "failed to begin transaction"))
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			}
		}()

		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO mtr_metrics (id, type, delta, value) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id, type) DO UPDATE
			SET delta = EXCLUDED.delta, value = EXCLUDED.value
		`)
		if err != nil {
			return backoff.Permanent(errors.Wrap(err, "failed to prepare upsert statement"))
		}
		defer stmt.Close()

		for _, metric := range *metrics {
			_, err := stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Delta, metric.Value)
			if err != nil {
				return backoff.Permanent(errors.Wrapf(err, "failed to upsert metric with ID %s", metric.ID))
			}
		}

		if err = tx.Commit(); err != nil {
			return backoff.Permanent(errors.Wrap(err, "failed to commit transaction"))
		}

		return nil
	})
}

// Ping checks the database connection with retry logic.
func (ds *DBStorage) Ping() error {
	if ds.Db == nil {
		return errors.New("db is nil")
	}

	return ds.Db.Ping()
}
