package router

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func SetRoutes(r *chi.Mux, handler func(http.ResponseWriter, *http.Request)) {
	r.Post("/update/{metricType}/{metricName}/{value}", handler)
}
