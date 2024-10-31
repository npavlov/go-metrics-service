package storage

import (
	"context"
	"time"

	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/repository"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// StorageMonitor monitors the database connection and provides the appropriate storage
type StorageMonitor struct {
	db         *gorm.DB
	memStorage *MemStorage
	dbStorage  *repository.DBRepository
	current    model.Repository
	logger     *zerolog.Logger
	ticker     *time.Ticker
}

// NewStorageMonitor initializes the storage monitor with database and memory storage options
func NewStorageMonitor(ctx context.Context, memStorage *MemStorage, dbStorage *repository.DBRepository, checkInterval time.Duration, log *zerolog.Logger) *StorageMonitor {

	monitor := &StorageMonitor{
		memStorage: memStorage,
		dbStorage:  dbStorage,
		logger:     log,
		ticker:     time.NewTicker(checkInterval),
	}

	// Check database connection and choose the appropriate storage at startup
	monitor.chooseStorage()

	// Start monitoring the database connection periodically
	go monitor.startMonitoring(ctx)

	return monitor
}

// chooseStorage selects dbStorage if the database is reachable; otherwise, it falls back to memStorage
func (s *StorageMonitor) chooseStorage() {
	if s.dbStorage.Ping() == nil {
		s.current = s.dbStorage
		s.logger.Info().Msg("Database is reachable. Using dbStorage as primary storage.")
	} else {
		s.current = s.memStorage
		s.logger.Warn().Msg("Database is unreachable. Using memStorage as fallback storage.")
	}
}

// startMonitoring periodically checks the database connection and updates the storage
func (s *StorageMonitor) startMonitoring(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.ticker.Stop()
			return
		case <-s.ticker.C:
			s.chooseStorage()
		}
	}
}

// GetRepository returns the current storage being used (either dbStorage or memStorage)
func (s *StorageMonitor) GetRepository() model.Repository {
	return s.current
}
