package storage

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	errNoValue = "no value provided"
)

type Number interface {
	int64 | float64
}

type Repository interface {
	Get(name domain.MetricName) (*model.Metric, bool)
	Create(metric *model.Metric) error
	GetAll() map[domain.MetricName]model.Metric
	Update(metric *model.Metric) error
	WithBackup(ctx context.Context, cfg *config.Config) *MemStorage
}

type MemStorage struct {
	mu      *sync.RWMutex
	metrics map[domain.MetricName]model.Metric
	cfg     *config.Config
	l       *zerolog.Logger
}

// NewMemStorage - constructor for MemStorage.
func NewMemStorage() *MemStorage {
	ms := &MemStorage{
		metrics: make(map[domain.MetricName]model.Metric),
		mu:      &sync.RWMutex{},
		l:       logger.NewLogger().Get(),
	}

	return ms
}

func (ms *MemStorage) WithBackup(ctx context.Context, cfg *config.Config) *MemStorage {
	ms.cfg = cfg

	err := ms.restore()
	if err != nil {
		// Continue on error
		ms.l.Error().Err(err).Msg("failed to restore metrics")
	}

	ms.l.Info().Msg("Metrics restored successfully")

	ms.startBackup(ctx)

	return ms
}

func (ms *MemStorage) startBackup(ctx context.Context) {
	if ms.cfg.StoreInterval > 0 {
		go func() {
			for {
				select {
				case <-ctx.Done():
					ms.l.Info().Msg("Stopping storage backup")

					return
				default:
					ms.mu.RLock()
					err := ms.saveFile()
					ms.mu.RUnlock()
					if err != nil {
						ms.l.Error().Err(err).Msg("Error saving file")
						panic(err)
					}
					time.Sleep(time.Duration(ms.cfg.StoreInterval) * time.Second)
				}
			}
		}()
	}
}

func (ms *MemStorage) restore() error {
	file, err := os.ReadFile(ms.cfg.File)
	if err != nil {
		ms.l.Error().Err(err).Msg("failed to load output")

		return errors.Wrap(err, "failed to load output")
	}
	newStorage := make(map[domain.MetricName]model.Metric)
	err = json.Unmarshal(file, &newStorage)
	if err != nil {
		ms.l.Error().Err(err).Msg("failed to unmarshal output")

		return errors.Wrap(err, "failed to unmarshal output")
	}

	ms.metrics = newStorage

	return nil
}

func (ms *MemStorage) saveFile() error {
	file, err := json.MarshalIndent(ms.metrics, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal output")
	}

	err = os.WriteFile(ms.cfg.File, file, fs.ModePerm)
	if err != nil {
		return errors.Wrap(err, "failed to save output")
	}

	return nil
}

func (ms *MemStorage) GetAll() map[domain.MetricName]model.Metric {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return cloneMap(ms.metrics)
}

// Get - retrieves the value of a Metric.
func (ms *MemStorage) Get(name domain.MetricName) (*model.Metric, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.metrics[name]

	return &value, exists
}

// Generic function to clone a map of Metrics.
func cloneMap(original map[domain.MetricName]model.Metric) map[domain.MetricName]model.Metric {
	cloned := make(map[domain.MetricName]model.Metric)
	for key, value := range original {
		cloned[key] = value
	}

	return cloned
}

func (ms *MemStorage) Update(metric *model.Metric) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if metric.Delta == nil && metric.Value == nil {
		return errors.New(errNoValue)
	}

	ms.metrics[metric.ID] = *metric

	if ms.cfg != nil && ms.cfg.StoreInterval == 0 {
		err := ms.saveFile()
		if err != nil {
			return err
		}
	}

	return nil
}

func (ms *MemStorage) Create(metric *model.Metric) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if metric.Delta == nil && metric.Value == nil {
		return errors.New(errNoValue)
	}

	ms.metrics[metric.ID] = *metric

	if ms.cfg != nil && ms.cfg.StoreInterval == 0 {
		err := ms.saveFile()
		if err != nil {
			return err
		}
	}

	return nil
}
