package main

import (
	"context"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

func main() {
	log := logger.NewLogger().SetLogLevel(zerolog.DebugLevel).Get()

	err := godotenv.Load("server.env")
	if err != nil {
		log.Error().Err(err).Msg("Error loading server.env file")
	}

	cfg := config.NewConfigBuilder(log).
		FromEnv().
		FromFlags().Build()

	log.Info().Interface("config", cfg).Msg("Configuration loaded")

	ctx, cancel := utils.WithSignalCancel(context.Background(), log)
	//nolint:exhaustruct
	db, err := gorm.Open(postgres.Open(cfg.Database), &gorm.Config{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database")
	}
	defer closeDB(db, log)

	memStorage := storage.NewMemStorage(log).WithBackup(ctx, cfg)
	// set repo to dbStorage if we are using database
	dbStorage := storage.NewDBStorage(db)

	stMonitor := storage.NewAutoSwitchRepo(memStorage, dbStorage, cfg.HealthCheckDur, log)
	stMonitor.StartMonitoring(ctx)

	mHandlers := handlers.NewMetricsHandler(stMonitor, log)
	hHandlers := handlers.NewHealthHandler(dbStorage, log)
	var cRouter router.Router = router.NewCustomRouter(log)
	cRouter.SetRouter(mHandlers, hHandlers)

	log.Info().
		Str("server_address", cfg.Address).
		Msg("Server started")

	//nolint:exhaustruct
	server := &http.Server{
		Addr:         cfg.Address,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		Handler:      cRouter.GetRouter(),
	}

	go func() {
		// Wait for the context to be done (i.e., signal received)
		<-ctx.Done()

		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Error shutting down server")
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Err(err).Msg("Error starting server")
		cancel()
	}

	log.Info().Msg("Server shut down")
}

// closeDB retrieves and closes the underlying sql.DB connection.
func closeDB(gormDB *gorm.DB, log *zerolog.Logger) {
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database")

		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close database connection")
	} else {
		log.Info().Msg("Database connection closed")
	}
}
