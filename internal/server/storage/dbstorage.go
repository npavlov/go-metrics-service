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
