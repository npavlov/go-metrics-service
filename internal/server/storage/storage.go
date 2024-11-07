package storage

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/server/snapshot"
)

const (
	errNoValue = "no value provided"
)

type MemStorage struct {
	mu       *sync.RWMutex
	metrics  map[domain.MetricName]db.Metric
	cfg      *config.Config
	l        *zerolog.Logger
	snapshot snapshot.Snapshot
}

// NewMemStorage - constructor for MemStorage.
func NewMemStorage(l *zerolog.Logger) *MemStorage {
	ms := &MemStorage{
		metrics:  make(map[domain.MetricName]db.Metric),
		mu:       &sync.RWMutex{},
		l:        l,
		cfg:      nil,
		snapshot: nil,
	}

	return ms
}

func (ms *MemStorage) WithBackup(ctx context.Context, cfg *config.Config) *MemStorage {
	// no file  provided
	if len(cfg.File) == 0 {
		ms.l.Warn().Msg("No file provided, running without backup")

		return ms
	}

	memSnapshot := snapshot.NewMemSnapshot(cfg.File, ms.l)
	ms.snapshot = memSnapshot
	ms.cfg = cfg

	if cfg.RestoreStorage {
		metrics, err := memSnapshot.Restore()
		if err != nil {
			ms.l.Error().Err(err).Msg("failed to restore metrics")
		}
		if err == nil {
			ms.metrics = metrics
			ms.l.Info().Msg("Metrics restored successfully")
		}
	}

	ms.StartBackup(ctx)
	ms.l.Info().Msgf("starting backup to file %s", cfg.File)

	return ms
}

func (ms *MemStorage) StartBackup(ctx context.Context) {
	if ms.cfg.StoreInterval > 0 {
		go func() {
			for {
				select {
				case <-ctx.Done():
					ms.l.Info().Msg("Stopping storage backup")
					ms.mu.RLock()
					_ = ms.snapshot.Save(ms.metrics)
					ms.mu.RUnlock()

					return
				default:
					ms.mu.RLock()
					err := ms.snapshot.Save(ms.metrics)
					ms.mu.RUnlock()
					if err != nil {
						ms.l.Error().Err(err).Msg("Error saving file")
						panic(err)
					}
					time.Sleep(ms.cfg.StoreIntervalDur)
				}
			}
		}()
	}
}

func (ms *MemStorage) GetAll(_ context.Context) map[domain.MetricName]db.Metric {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return cloneMap(ms.metrics)
}

// Get - retrieves the value of a Metric.
func (ms *MemStorage) Get(_ context.Context, name domain.MetricName) (*db.Metric, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.metrics[name]

	return &value, exists
}

// GetMany retrieves multiple metrics by their names.
func (ms *MemStorage) GetMany(_ context.Context, names []domain.MetricName) (map[domain.MetricName]db.Metric, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	results := make(map[domain.MetricName]db.Metric)
	for _, name := range names {
		metric, exists := ms.metrics[name]
		if exists {
			results[metric.ID] = metric
		}
	}

	return results, nil
}

// Generic function to clone a map of Metrics.
func cloneMap(original map[domain.MetricName]db.Metric) map[domain.MetricName]db.Metric {
	cloned := make(map[domain.MetricName]db.Metric)
	for key, value := range original {
		cloned[key] = value
	}

	return cloned
}

func (ms *MemStorage) Update(_ context.Context, metric *db.Metric) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if metric.Delta == nil && metric.Value == nil {
		return errors.New(errNoValue)
	}

	ms.metrics[metric.ID] = *metric

	if ms.cfg != nil && ms.cfg.StoreInterval == 0 {
		err := ms.snapshot.Save(ms.metrics)
		if err != nil {
			return errors.Wrap(err, "failed to save metrics")
		}
	}

	return nil
}

func (ms *MemStorage) Create(_ context.Context, metric *db.Metric) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if metric.Delta == nil && metric.Value == nil {
		return errors.New(errNoValue)
	}

	ms.metrics[metric.ID] = *metric

	if ms.cfg != nil && ms.cfg.StoreInterval == 0 {
		err := ms.snapshot.Save(ms.metrics)
		if err != nil {
			return errors.Wrap(err, "failed to save metrics")
		}
	}

	return nil
}

func (ms *MemStorage) UpdateMany(_ context.Context, metrics *[]db.Metric) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, metric := range *metrics {
		ms.metrics[metric.ID] = metric
	}

	return nil
}
