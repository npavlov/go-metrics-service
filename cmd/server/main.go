//nolint:gochecknoglobals,ireturn,lll
package main

import (
	"context"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/dbmanager"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	log := setupLogger()
	cfg := loadConfig(&log)

	ctx, cancel := utils.WithSignalCancel(context.Background(), &log)
	defer cancel()

	dbManager := setupDBManager(ctx, cfg, &log)
	defer dbManager.Close()

	metricStorage := setupStorage(ctx, cfg, dbManager, &log)

	startServer(ctx, cfg, metricStorage, dbManager, &log)
}

func setupLogger() zerolog.Logger {
	log := logger.NewLogger(zerolog.DebugLevel).Get()

	log.Info().Str("buildVersion", buildVersion).
		Str("buildCommit", buildCommit).
		Str("buildDate", buildDate).Msg("Starting server")

	return log
}

func loadConfig(log *zerolog.Logger) *config.Config {
	if err := godotenv.Load("server.env"); err != nil {
		log.Error().Err(err).Msg("Error loading server.env file")
	}

	cfg := config.NewConfigBuilder(log).
		FromEnv().
		FromFlags().Build()

	log.Info().Interface("config", cfg).Msg("Configuration loaded")

	return cfg
}

func setupDBManager(ctx context.Context, cfg *config.Config, log *zerolog.Logger) *dbmanager.DBManager {
	dbManager := dbmanager.NewDBManager(cfg.Database, log).Connect(ctx).ApplyMigrations()

	if !dbManager.IsConnected {
		log.Error().Msg("Database connection failed, using in-memory storage")
	}

	return dbManager
}

func setupStorage(ctx context.Context, cfg *config.Config, dbManager *dbmanager.DBManager, log *zerolog.Logger) model.Repository {
	if dbManager.IsConnected {
		return storage.NewDBStorage(dbManager.DB, log)
	}

	return storage.NewMemStorage(log).WithBackup(ctx, cfg)
}

func startServer(ctx context.Context, cfg *config.Config, metricStorage model.Repository, dbManager *dbmanager.DBManager, log *zerolog.Logger) {
	mHandlers := handlers.NewMetricsHandler(metricStorage, log)
	hHandlers := handlers.NewHealthHandler(dbManager, log)

	cRouter := router.NewCustomRouter(cfg, log)
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
		<-ctx.Done()
		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Error shutting down server")
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Err(err).Msg("Error starting server")
	}
	log.Info().Msg("Server shut down")
}
