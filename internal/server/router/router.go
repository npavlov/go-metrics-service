package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/npavlov/go-metrics-service/internal/types"
)

func GetRouter(handlers types.Handlers) *chi.Mux {

	// Create a new chi router
	r := chi.NewRouter()

	// Useful middlewares, extra logging
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.RenderHandler)
		r.Route("/update", func(r chi.Router) {
			r.Post("/{metricType}/{metricName}/{value}", handlers.UpdateHandler)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{metricType}/{metricName}", handlers.RetrieveHandler)
		})
	})

	return r
}
