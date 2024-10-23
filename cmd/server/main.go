package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/npavlov/go-metrics-service/internal/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func main() {
	log := logger.NewLogger().SetLogLevel(zerolog.DebugLevel).Get()

	err := godotenv.Load("server.env")
	if err != nil {
		log.Fatal().Msg("Error loading server.env file")
	}

	cfg := config.NewConfigBuilder(log).
		FromEnv().
		FromFlags().Build()

	log.Info().Interface("config", cfg).Msg("Configuration loaded")

	ctx := utils.WithSignalCancel(context.Background(), log)

	var memStorage storage.Repository = storage.NewMemStorage(log).WithBackup(ctx, cfg)

	router := chi.NewRouter()
	var mHandlers handlers.Handlers = handlers.NewMetricsHandler(memStorage, router, log)
	mHandlers.SetRouter()

	log.Info().
		Str("server_address", cfg.Address).
		Msg("Server started")

	//nolint:exhaustruct
	server := &http.Server{
		Addr:         cfg.Address,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		Handler:      router,
	}

	go func() {
		// Wait for the context to be done (i.e., signal received)
		<-ctx.Done()

		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Error shutting down server")
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("Error starting server")
	}

	log.Info().Msg("Server shut down")
}
