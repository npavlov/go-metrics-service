package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/npavlov/go-metrics-service/internal/server/handlers/render"
	"github.com/npavlov/go-metrics-service/internal/server/handlers/retrieve"
	"github.com/npavlov/go-metrics-service/internal/server/handlers/update"
	"github.com/npavlov/go-metrics-service/internal/storage"
)

func GetRouter(memStorage storage.Repository) *chi.Mux {

	// Create a new chi router
	r := chi.NewRouter()

	// Useful middlewares, extra logging
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	updateHandler := update.GetUpdateHandler(memStorage)
	retrieveHandle := retrieve.GetRetrieveHandler(memStorage)
	renderHandler := render.GetRenderHandler(memStorage)

	r.Route("/", func(r chi.Router) {
		r.Get("/", renderHandler)
		r.Route("/update", func(r chi.Router) {
			r.Post("/{metricType}/{metricName}/{value}", updateHandler)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{metricType}/{metricName}", retrieveHandle)
		})
	})

	return r
}
