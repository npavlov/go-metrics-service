package storage

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
)

// AutoSwitchRepo monitors the database connection and provides the appropriate storage.
type AutoSwitchRepo struct {
	memStorage *MemStorage
	dbStorage  *DBStorage
	current    model.Repository
	logger     *zerolog.Logger
	ticker     *time.Ticker
}

// NewAutoSwitchRepo initializes the storage monitor with database and memory storage options.
func NewAutoSwitchRepo(mems *MemStorage, dbs *DBStorage, interval time.Duration, log *zerolog.Logger) *AutoSwitchRepo {
	monitor := &AutoSwitchRepo{
		memStorage: mems,
		dbStorage:  dbs,
		logger:     log,
		ticker:     time.NewTicker(interval),
		current:    mems,
	}

	// Check database connection and choose the appropriate storage at startup
	monitor.chooseStorage()

	return monitor
}

func (awr *AutoSwitchRepo) StartMonitoring(ctx context.Context) {
	// Start monitoring the database connection periodically
	go awr.startMonitoring(ctx)
}

// chooseStorage selects dbStorage if the database is reachable; otherwise, it falls back to memStorage.
func (awr *AutoSwitchRepo) chooseStorage() {
	if awr.dbStorage.Ping() == nil {
		awr.current = awr.dbStorage
		awr.logger.Info().Msg("Database is reachable. Using dbStorage as primary storage.")
	} else {
		awr.current = awr.memStorage
		awr.logger.Warn().Msg("Database is unreachable. Using memStorage as fallback storage.")
	}
}

// startMonitoring periodically checks the database connection and updates the storage.
func (awr *AutoSwitchRepo) startMonitoring(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			awr.ticker.Stop()

			return
		case <-awr.ticker.C:
			awr.chooseStorage()
		}
	}
}

func (awr *AutoSwitchRepo) GetAll(ctx context.Context) map[domain.MetricName]model.Metric {
	return awr.current.GetAll(ctx)
}

// Get - retrieves the value of a Metric.
func (awr *AutoSwitchRepo) Get(ctx context.Context, name domain.MetricName) (*model.Metric, bool) {
	return awr.current.Get(ctx, name)
}

func (awr *AutoSwitchRepo) Update(ctx context.Context, metric *model.Metric) error {
	err := awr.current.Update(ctx, metric)
	if err != nil {
		return errors.Wrap(err, "failed to update metric")
	}

	return nil
}

func (awr *AutoSwitchRepo) Create(ctx context.Context, metric *model.Metric) error {
	err := awr.current.Create(ctx, metric)
	if err != nil {
		return errors.Wrap(err, "could not create metric")
	}

	return nil
}
