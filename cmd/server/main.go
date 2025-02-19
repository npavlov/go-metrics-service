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
	"github.com/npavlov/go-metrics-service/internal/server/buildinfo"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/dbmanager"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

func main() {
	log := logger.NewLogger(zerolog.DebugLevel).Get()

	log.Info().Str("buildVersion", buildinfo.Version).
		Str("buildCommit", buildinfo.Commit).
		Str("buildDate", buildinfo.Date).Msg("Starting server")

	cfg := loadConfig(&log)

	ctx, cancel := utils.WithSignalCancel(context.Background(), &log)
	defer cancel()

	dbManager := dbmanager.NewDBManager(cfg.Database, &log).Connect(ctx).ApplyMigrations()
	defer dbManager.Close()

	var metricStorage model.Repository
	if dbManager.IsConnected {
		metricStorage = storage.NewDBStorage(dbManager.DB, &log)
	} else {
		metricStorage = storage.NewMemStorage(&log).WithBackup(ctx, cfg)
	}

	startServer(ctx, cfg, metricStorage, dbManager, &log)
}

func loadConfig(log *zerolog.Logger) *config.Config {
	if err := godotenv.Load("server.env"); err != nil {
		log.Error().Err(err).Msg("Error loading server.env file")
	}

	cfg := config.NewConfigBuilder(log).
		FromEnv().
		FromFlags().
		FromFile().Build()

	log.Info().Interface("config", cfg).Msg("Configuration loaded")

	return cfg
}

func startServer(
	ctx context.Context,
	cfg *config.Config,
	metricStorage model.Repository,
	dbManager *dbmanager.DBManager,
	log *zerolog.Logger,
) {
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
