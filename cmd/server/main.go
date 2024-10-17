package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
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

	var memStorage storage.Repository = storage.NewMemStorage()
	r := chi.NewRouter()
	var router handlers.Handlers = handlers.NewMetricsHandler(memStorage, r)
	router.SetRouter()

	// Launching server at :8080
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
	err = server.ListenAndServe()
	if err != nil {
		l.Fatal().Err(err).Msg("Error starting server")
	}
}
