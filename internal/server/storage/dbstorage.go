package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
)

type DBStorage struct {
	db *gorm.DB
}

func NewDBStorage(db *gorm.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

func retryOperation(operation func() error) error {
	maxRetries := 3
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for idx := 0; idx <= maxRetries; idx++ {
		err := operation()
		if err == nil {
			return nil
		}

		// Проверяем, является ли ошибка retriable
		if !isRetriableError(err) {
			return err
		}

		// Ждём перед следующей попыткой
		if idx < maxRetries {
			time.Sleep(retryDelays[idx])
		}
	}

	return errors.New("operation failed after retries")
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException, pgerrcode.ConnectionFailure, pgerrcode.CrashShutdown:
			return true
		}
	}

	return false
}

func (repo *DBStorage) Ping() error {
	if repo.db == nil {
		return errors.New("db is nil")
	}

	db, err := repo.db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	return errors.Wrap(db.Ping(), "failed to ping DB")
}

// Create inserts a new metric into the database.
func (repo *DBStorage) Create(context context.Context, metric *model.Metric) error {
	return retryOperation(func() error {
		if err := repo.db.WithContext(context).Create(metric).Error; err != nil {
			return fmt.Errorf("failed to create metric: %w", err)
		}

		return nil
	})
}

// Get fetches a single metric by name.
func (repo *DBStorage) Get(context context.Context, name domain.MetricName) (*model.Metric, bool) {
	var metric model.Metric
	err := retryOperation(func() error {
		return repo.db.WithContext(context).First(&metric, "id = ?", name).Error
	})
	if err != nil {
		return nil, false
	}

	return &metric, true
}

// GetMany fetches a single metric by name.
//
//nolint:lll
func (repo *DBStorage) GetMany(context context.Context, names []domain.MetricName) (map[domain.MetricName]model.Metric, error) {
	var metrics []model.Metric
	err := retryOperation(func() error {
		return repo.db.WithContext(context).Where("id IN ?", names).Find(&metrics).Error
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metrics")
	}

	results := make(map[domain.MetricName]model.Metric)

	for _, metric := range metrics {
		results[metric.ID] = metric
	}

	return results, nil
}

// GetAll fetches all metrics from the database.
func (repo *DBStorage) GetAll(context context.Context) map[domain.MetricName]model.Metric {
	var metrics []model.Metric
	result := make(map[domain.MetricName]model.Metric)

	if err := repo.db.WithContext(context).Find(&metrics).Error; err != nil {
		return result
	}

	for _, metric := range metrics {
		result[metric.ID] = metric
	}

	return result
}

// Update modifies an existing metric in the database.
func (repo *DBStorage) Update(context context.Context, metric *model.Metric) error {
	return retryOperation(func() error {
		if err := repo.db.WithContext(context).Save(metric).Error; err != nil {
			return fmt.Errorf("failed to update metric: %w", err)
		}

		return nil
	})
}

// UpdateMany inserts or updates multiple metrics in the database.
func (repo *DBStorage) UpdateMany(ctx context.Context, metrics *[]model.Metric) error {
	return retryOperation(func() error {
		tx := repo.db.WithContext(ctx).Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to start transaction: %w", tx.Error)
		}

		for _, metric := range *metrics {
			if err := tx.Save(&metric).Error; err != nil {
				tx.Rollback()

				return fmt.Errorf("failed to save metric with ID %s: %w", metric.ID, err)
			}
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})
}
