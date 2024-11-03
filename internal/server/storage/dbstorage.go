package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type DBStorage struct {
	db *sql.DB
	l  *zerolog.Logger
}

// NewDBStorage initializes a new DBStorage instance.
func NewDBStorage(db *sql.DB, l *zerolog.Logger) *DBStorage {
	gg := db.Ping()
	if gg != nil {
		l.Error().Msg("TEGGE")
	}

	return &DBStorage{
		db: db,
		l:  l,
	}
}

// GetAll retrieves all metrics from the database.
func (ds *DBStorage) GetAll(ctx context.Context) map[domain.MetricName]model.Metric {
	query := `SELECT id, type, delta, value FROM public.mtr_metrics`
	rows, err := ds.db.QueryContext(ctx, query)
	if err != nil {
		ds.l.Error().Err(err).Msg("error getting metrics")
		return nil
	}
	defer rows.Close()

	metrics := make(map[domain.MetricName]model.Metric)
	for rows.Next() {
		var metric model.Metric
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
			ds.l.Error().Err(err).Msg("error getting metrics")
			return nil
		}
		metrics[metric.ID] = metric
	}

	if err := rows.Err(); err != nil {
		ds.l.Error().Err(err).Msg("error getting metrics")
		return nil
	}

	return metrics
}

// Get retrieves a single metric by its name.
func (ds *DBStorage) Get(ctx context.Context, name domain.MetricName) (*model.Metric, bool) {
	query := `SELECT id, type, delta, value FROM public.mtr_metrics WHERE id = $1`
	row := ds.db.QueryRowContext(ctx, query, name)

	var metric model.Metric
	if err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
		ds.l.Error().Err(err).Msg("failed to retrieve metric")
		return nil, false
	}

	return &metric, true
}

// GetMany retrieves multiple metrics based on their names.
func (ds *DBStorage) GetMany(ctx context.Context, names []domain.MetricName) (map[domain.MetricName]model.Metric, error) {
	// Return an empty result if no names are provided
	if len(names) == 0 {
		return map[domain.MetricName]model.Metric{}, nil
	}

	// Create placeholders and arguments for each name in the slice
	placeholders := make([]string, len(names))
	args := make([]interface{}, len(names))
	for i, name := range names {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = name
	}

	// Build the SQL query with placeholders
	query := fmt.Sprintf(`SELECT id, type, delta, value FROM public.mtr_metrics WHERE id IN (%s)`, strings.Join(placeholders, ", "))

	// Execute the query
	rows, err := ds.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve multiple metrics")
	}
	defer rows.Close()

	// Prepare a map to hold the results
	metrics := make(map[domain.MetricName]model.Metric)
	for rows.Next() {
		var metric model.Metric
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
			return nil, errors.Wrap(err, "failed to scan metric row")
		}
		metrics[metric.ID] = metric
	}

	// Check for errors that might have occurred during row iteration
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error occurred during row iteration")
	}

	return metrics, nil
}

// Update modifies an existing metric in the database.
func (ds *DBStorage) Update(ctx context.Context, metric *model.Metric) error {
	query := `
		UPDATE public.mtr_metrics 
		SET delta = $2, value = $3 
		WHERE id = $1 AND type = $4`
	_, err := ds.db.ExecContext(ctx, query, metric.ID, metric.Delta, metric.Value, metric.MType)
	if err != nil {
		return errors.Wrap(err, "failed to update metric")
	}
	return nil
}

// Create inserts a new metric into the database.
func (ds *DBStorage) Create(ctx context.Context, metric *model.Metric) error {
	query := `
		INSERT INTO public.mtr_metrics (id, type, delta, value) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id, type) DO NOTHING`
	_, err := ds.db.ExecContext(ctx, query, metric.ID, metric.MType, metric.Delta, metric.Value)
	if err != nil {
		ds.l.Error().Err(err).Msg("failed to insert metric")
		return errors.Wrap(err, "failed to create metric")
	}
	return nil
}

// UpdateMany updates multiple metrics in the database.
func (ds *DBStorage) UpdateMany(ctx context.Context, metrics *[]model.Metric) error {
	tx, err := ds.db.BeginTx(ctx, nil) // Start a transaction
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback() // Rollback if there's an error
		}
	}()

	// Prepare an upsert statement using ON CONFLICT
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO public.mtr_metrics (id, type, delta, value) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id, type) DO UPDATE
		SET delta = EXCLUDED.delta, value = EXCLUDED.value
	`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare upsert statement")
	}
	defer stmt.Close()

	// Execute the upsert statement for each metric in the list
	for _, metric := range *metrics {
		_, err := stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Delta, metric.Value)
		if err != nil {
			return errors.Wrapf(err, "failed to upsert metric with ID %s", metric.ID)
		}
	}

	// Commit the transaction if everything is successful
	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (ds *DBStorage) Ping() error {
	if ds.db == nil {
		return errors.New("db is nil")
	}

	return ds.db.Ping()
}
