package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
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
	log     *zerolog.Logger
	dbCon   dbmanager.PgxPool
	builder squirrel.StatementBuilderType
}

// NewDBStorage initializes a new DBStorage instance.
func NewDBStorage(dbCon dbmanager.PgxPool, log *zerolog.Logger) *DBStorage {
	return &DBStorage{
		dbCon:   dbCon,
		log:     log,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

type OperationFunc func() error

// retryOperation executes a database operation with retry logic.
func (ds *DBStorage) retryOperation(ctx context.Context, operation OperationFunc) error {
	backoffConfig := backoff.NewExponentialBackOff()
	backoffConfig.InitialInterval = 1 * time.Second
	backoffConfig.Multiplier = 3
	retryWithLimit := backoff.WithMaxRetries(backoffConfig, maxRetries)

	err := backoff.Retry(func() error {
		err := operation()
		if err != nil && ds.isRetryableError(err) {
			ds.log.Warn().Err(err).Msg("transient error, retrying operation")

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

	query := ds.builder.
		Select("m.id", "m.type", "c.delta", "g.value").
		From("mtr_metrics AS m").
		LeftJoin("counter_metrics AS c ON m.id = c.metric_id").
		LeftJoin("gauge_metrics AS g ON m.id = g.metric_id")

	sql, args, err := query.ToSql()
	if err != nil {
		return metrics
	}

	err = ds.retryOperation(ctx, func() error {
		rows, err := ds.dbCon.Query(ctx, sql, args...)
		defer rows.Close()
		if err != nil {
			ds.log.Error().Err(err).Msg("error getting metrics")

			return errors.Wrap(err, "error getting metrics")
		}

		for rows.Next() {
			metric := db.Metric{}

			err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
			if err != nil {
				return errors.Wrap(err, "failed to scan GetAll row")
			}

			metrics[metric.ID] = metric
		}

		return nil
	})
	if err != nil {
		ds.log.Error().Err(err).Msg("failed to retrieve metrics after retries")

		return nil
	}

	return metrics
}

// Get retrieves a single metric by its name with retry logic.
func (ds *DBStorage) Get(ctx context.Context, name domain.MetricName) (*db.Metric, bool) {
	metric := db.Metric{}

	query := ds.builder.
		Select("m.id", "m.type", "c.delta", "g.value").
		From("mtr_metrics AS m").
		LeftJoin("counter_metrics AS c ON m.id = c.metric_id").
		LeftJoin("gauge_metrics AS g ON m.id = g.metric_id").
		Where(squirrel.Eq{"m.id": name})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, false
	}

	err = ds.retryOperation(ctx, func() error {
		row := ds.dbCon.QueryRow(ctx, sql, args...)

		err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			return errors.Wrap(err, "failed to scan GetAll row")
		}

		return nil
	})
	if err != nil {
		return nil, false
	}

	return &metric, true
}

func (ds *DBStorage) GetMany(ctx context.Context, names []domain.MetricName) (map[domain.MetricName]db.Metric, error) {
	if len(names) == 0 {
		return map[domain.MetricName]db.Metric{}, nil
	}

	metrics := make(map[domain.MetricName]db.Metric)

	err := ds.retryOperation(ctx, func() error {
		// Build the query with Squirrel
		query, args, err := ds.builder.Select(
			"m.id",
			"m.type",
			"c.delta",
			"g.value",
		).
			From("mtr_metrics AS m").
			LeftJoin("counter_metrics AS c ON m.id = c.metric_id").
			LeftJoin("gauge_metrics AS g ON m.id = g.metric_id").
			Where(squirrel.Eq{"m.id": names}). // Use Squirrel's In clause
			ToSql()
		if err != nil {
			ds.log.Error().Err(err).Msg("failed to build GetMany query")
			return errors.Wrap(err, "failed to build GetMany query")
		}

		// Execute the query
		rows, err := ds.dbCon.Query(ctx, query, args...)
		if err != nil {
			ds.log.Error().Err(err).Msg("error executing GetMany query")
			return errors.Wrap(err, "error executing GetMany query")
		}
		defer rows.Close()

		// Parse the results
		for rows.Next() {
			metric := db.Metric{}

			if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
				ds.log.Error().Err(err).Msg("error scanning row for GetMany")
				return errors.Wrap(err, "error scanning row for GetMany")
			}

			metrics[metric.ID] = metric
		}

		if err := rows.Err(); err != nil {
			ds.log.Error().Err(err).Msg("error iterating over GetMany results")
			return errors.Wrap(err, "error iterating over GetMany results")
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
		var query squirrel.UpdateBuilder

		switch metric.MType {
		case domain.Gauge:
			query = ds.builder.Update("gauge_metrics").
				Set("value", metric.Value).
				Where(squirrel.Eq{"metric_id": metric.ID})
		case domain.Counter:
			query = ds.builder.Update("counter_metrics").
				Set("delta", metric.Delta).
				Where(squirrel.Eq{"metric_id": metric.ID})
		}

		// Construct the SQL query with the necessary parameters.
		sql, args, err := query.ToSql()
		if err != nil {
			return errors.Wrap(err, "failed to build SQL query")
		}

		// Execute the query using the database connection.
		_, err = ds.dbCon.Exec(ctx, sql, args...)
		if err != nil {
			ds.log.Error().Err(err).Msg("error updating metric")
			return errors.Wrap(err, "error updating metric")
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

	// Retry operation for database transaction
	err := ds.retryOperation(ctx, func() error {
		// Begin a new transaction
		tx, err := ds.dbCon.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			ds.log.Error().Err(err).Msg("failed to start transaction for Create")
			return errors.Wrap(err, "failed to start transaction for Create")
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback(ctx)
			} else {
				err = tx.Commit(ctx)
			}
		}()

		// Insert into mtr_metrics table
		insertMetricSQL, metricArgs, err := ds.builder.
			Insert("mtr_metrics").
			Columns("id", "type").
			Values(metric.ID, metric.MType).
			Suffix("ON CONFLICT (id, type) DO NOTHING").
			ToSql()
		if err != nil {
			ds.log.Error().Err(err).Msg("failed to build mtr_metrics insert query")
			return errors.Wrap(err, "failed to build mtr_metrics insert query")
		}

		_, err = tx.Exec(ctx, insertMetricSQL, metricArgs...)
		if err != nil {
			ds.log.Error().Err(err).Msg("failed to insert metric")
			return errors.Wrap(err, "failed to insert metric")
		}

		var query squirrel.InsertBuilder

		// Insert into counter_metrics or gauge_metrics table based on metric type
		switch metric.MType {
		case domain.Gauge:
			query = ds.builder.
				Insert("gauge_metrics").
				Columns("metric_id", "value").
				Values(metric.ID, metric.Value).
				Suffix("ON CONFLICT (metric_id) DO NOTHING")

		case domain.Counter:
			query = ds.builder.
				Insert("counter_metrics").
				Columns("metric_id", "delta").
				Values(metric.ID, metric.Delta).
				Suffix("ON CONFLICT (metric_id) DO NOTHING")
		}

		insertCounterSQL, metricArgs, err := query.ToSql()

		if err != nil {
			ds.log.Error().Err(err).Msg("failed to build mtr_metrics insert query")
			return errors.Wrap(err, "failed to build mtr_metrics insert query")
		}

		_, err = tx.Exec(ctx, insertCounterSQL, metricArgs...)
		if err != nil {
			ds.log.Error().Err(err).Msg("failed to insert metric")
			return errors.Wrap(err, "failed to insert metric")
		}

		return nil
	})

	if err != nil {
		ds.log.Error().Err(err).Msg("failed to create metric")
		return errors.Wrap(err, "failed to create metric")
	}

	return nil
}

// UpdateMany updates multiple metrics in the database with retry logic.
func (ds *DBStorage) UpdateMany(ctx context.Context, metrics *[]db.Metric) error {
	return ds.retryOperation(ctx, func() error {
		err := WithTx(ctx, ds.dbCon, func(ctx context.Context, tx pgx.Tx) error {
			key1, key2 := KeyNameAsHash64("update_many")
			err := AcquireBlockingLock(ctx, tx, key1, key2, ds.log)
			if err != nil {
				ds.log.Error().Err(err).Msg("failed to acquire lock")

				return errors.Wrap(err, "failed to acquire lock")
			}

			for _, metric := range *metrics {
				upsertMetricQuery := ds.builder.Insert("mtr_metrics").
					Columns("id", "type").
					Values(metric.ID, metric.MType).
					Suffix("ON CONFLICT (id) DO UPDATE SET type = EXCLUDED.type")
				sql, args, err := upsertMetricQuery.ToSql()
				if err != nil {
					ds.log.Error().Err(err).Msg("failed to build UPSERT query for metric")
					return errors.Wrap(err, "failed to build UPSERT query for metric")
				}

				// Execute the UPSERT for the main metric record.
				_, err = tx.Exec(ctx, sql, args...)
				if err != nil {
					ds.log.Error().Err(err).Msg("failed to upsert metric")
					return errors.Wrap(err, "failed to upsert metric")
				}

				var query squirrel.InsertBuilder

				switch metric.MType {
				case domain.Gauge:
					query = ds.builder.Insert("gauge_metrics").
						Columns("value", "metric_id").
						Values(metric.Value, metric.ID).
						Suffix("ON CONFLICT (metric_id) DO UPDATE SET value = EXCLUDED.value")
				case domain.Counter:
					query = ds.builder.Insert("counter_metrics").
						Columns("delta", "metric_id").
						Values(metric.Delta, metric.ID).
						Suffix("ON CONFLICT (metric_id) DO UPDATE SET delta = EXCLUDED.delta")
				}

				sql, args, err = query.ToSql()
				if err != nil {
					ds.log.Error().Err(err).Msg("failed to build UPSERT query for counter")
					return errors.Wrap(err, "failed to build UPSERT query for counter")
				}

				// Execute the UPSERT for the Counter metric.
				_, err = tx.Exec(ctx, sql, args...)
				if err != nil {
					ds.log.Error().Err(err).Msg("failed to upsert counter metric")
					return errors.Wrap(err, "failed to upsert counter metric")
				}
			}

			return nil
		})
		if err != nil {
			ds.log.Error().Err(err).Msg("error in UpdateMany")

			return errors.Wrap(err, "error in UpdateMany")
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
