package main

import (
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/flags"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"log"
	"net/http"
)

func main() {
	cfg := parseFlags()
	flags.VerifyFlags()

	var memStorage storage.Repository = storage.NewMemStorage()

	r := router.GetRouter(memStorage)

	// Launching server at :8080
	fmt.Printf("Server started at %s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
