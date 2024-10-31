package storage

import (
	"context"
	"fmt"

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
	if err := repo.db.WithContext(context).Create(metric).Error; err != nil {
		return fmt.Errorf("failed to create metric: %w", err)
	}

	return nil
}

// Get fetches a single metric by name.
func (repo *DBStorage) Get(context context.Context, name domain.MetricName) (*model.Metric, bool) {
	var metric model.Metric
	if err := repo.db.WithContext(context).First(&metric, "id = ?", name).Error; err != nil {
		return nil, false
	}

	return &metric, true
}

// GetMany fetches a single metric by name.
//
//nolint:lll
func (repo *DBStorage) GetMany(context context.Context, names []domain.MetricName) (*map[domain.MetricName]model.Metric, error) {
	var metrics []model.Metric
	if err := repo.db.WithContext(context).Where("id IN ?", names).Find(&metrics).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get metrics")
	}

	results := make(map[domain.MetricName]model.Metric)

	for _, metric := range metrics {
		results[metric.ID] = metric
	}

	return &results, nil
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
	if err := repo.db.WithContext(context).Save(metric).Error; err != nil {
		return fmt.Errorf("failed to update metric: %w", err)
	}

	return nil
}

// UpdateMany inserts or updates multiple metrics in the database.
func (repo *DBStorage) UpdateMany(ctx context.Context, metrics *[]model.Metric) error {
	// Start a transaction
	tx := repo.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Attempt to save each metric
	for _, metric := range *metrics {
		if err := tx.Save(&metric).Error; err != nil {
			// Rollback transaction if any save fails
			tx.Rollback()

			return fmt.Errorf("failed to save metric with ID %s: %w", metric.ID, err)
		}
	}

	// Commit the transaction if all saves are successful
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
