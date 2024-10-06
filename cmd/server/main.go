package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"log"
	"net/http"
)

func main() {
	cfg := config.NewConfigBuilder().
		FromEnv().
		FromFlags().Build()

	var memStorage storage.Repository = storage.NewMemStorage()
	var r = chi.NewRouter()
	handlers.NewMetricsHandler(memStorage, r)

	// Launching server at :8080
	fmt.Printf("Server started at %s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
