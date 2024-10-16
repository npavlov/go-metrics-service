package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"net/http"
)

func main() {
	l := logger.Get()
	logger.SetLogLevel()

	err := godotenv.Load("server.env")
	if err != nil {
		l.Fatal().Msg("Error loading server.env file")
	}

	cfg := config.NewConfigBuilder().
		FromEnv().
		FromFlags().Build()

	var memStorage storage.Repository = storage.NewMemStorage()
	var r = chi.NewRouter()
	handlers.NewMetricsHandler(memStorage, r)

	// Launching server at :8080
	l.Info().
		Str("server_address", cfg.Address).
		Msg("Server started")
	err = http.ListenAndServe(cfg.Address, r)
	if err != nil {
		l.Fatal().Err(err).Msg("Error starting server")
	}
}
