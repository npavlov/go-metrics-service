package main

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/repository"
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

	var memStorage storage.InMemory = storage.NewMemStorage(log).WithBackup(ctx, cfg)

	dbpool, err := pgxpool.New(context.Background(), cfg.Database)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database")
	}
	defer dbpool.Close()

	dbrepository := repository.NewDBRepository(dbpool)

	universal := repository.NewUniversal(dbrepository, memStorage)

	mHandlers := handlers.NewMetricsHandler(*universal, log)
	var cRouter router.Router = router.NewCustomRouter(log)
	cRouter.SetRouter(mHandlers)

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
