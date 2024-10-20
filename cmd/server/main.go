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
	l := logger.NewLogger().SetLogLevel(zerolog.DebugLevel).Get()

	err := godotenv.Load("server.env")
	if err != nil {
		l.Fatal().Msg("Error loading server.env file")
	}

	cfg := config.NewConfigBuilder().
		FromEnv().
		FromFlags().Build()

	l.Info().Interface("config", cfg).Msg("Configuration loaded")

	ctx := utils.WithSignalCancel(context.Background())

	var memStorage storage.Repository = storage.NewMemStorage().WithBackup(ctx, cfg)

	r := chi.NewRouter()
	var router handlers.Handlers = handlers.NewMetricsHandler(memStorage, r)
	router.SetRouter()

	l.Info().
		Str("server_address", cfg.Address).
		Msg("Server started")

	//nolint:exhaustruct
	server := &http.Server{
		Addr:         cfg.Address,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		Handler:      r,
	}

	go func() {
		// Wait for the context to be done (i.e., signal received)
		<-ctx.Done()

		if err := server.Shutdown(ctx); err != nil {
			l.Error().Err(err).Msg("Error shutting down server")
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Fatal().Err(err).Msg("Error starting server")
	}

	l.Info().Msg("Server shut down")
}
