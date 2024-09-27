package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/npavlov/go-metrics-service/internal/server/handler"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"log"
	"net/http"
)

func main() {
	var memStorage storage.Repository = storage.NewMemStorage()

	serverHandler := handler.GetUpdateHandler(memStorage)

	// Create a new chi router
	r := chi.NewRouter()

	// Useful middlewares, extra logging
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	router.SetRoutes(r, serverHandler)

	// Launching server at :8080
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
