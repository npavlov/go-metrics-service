package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
	"github.com/npavlov/go-metrics-service/pkg/crypto"
)

const (
	defaultTimeout = 500 * time.Millisecond // Default timeout for metrics handler
)

type Router interface {
	SetRouter(mh *handlers.MetricHandler, hh *handlers.HealthHandler)
	GetRouter() *chi.Mux
}

type CustomRouter struct {
	router     *chi.Mux
	logger     *zerolog.Logger
	cfg        *config.Config
	decryption *crypto.Decryption
}

// NewCustomRouter - constructor for CustomRouter.
func NewCustomRouter(cfg *config.Config, log *zerolog.Logger) *CustomRouter {
	router := &CustomRouter{
		router:     chi.NewRouter(),
		logger:     log,
		cfg:        cfg,
		decryption: nil,
	}

	if cfg.CryptoKey == "" {
		return router
	}

	decryption, err := crypto.NewDecryption(cfg.CryptoKey)
	if err != nil {
		log.Fatal().Err(err).Msg("could not create decryption")
	}

	router.decryption = decryption

	return router
}

// SetRouter Embedding middleware setup in the constructor.
func (cr *CustomRouter) SetRouter(mh *handlers.MetricHandler, hh *handlers.HealthHandler) {
	cr.router.Use(middlewares.LoggingMiddleware(cr.logger))
	cr.router.Use(middlewares.TimeoutMiddleware(defaultTimeout))
	cr.router.Use(middleware.Recoverer)
	cr.router.Use(middlewares.GzipMiddleware)
	cr.router.Use(middlewares.BrotliMiddleware)
	if cr.decryption != nil {
		cr.router.Use(middlewares.DecryptMiddleware(cr.decryption, cr.logger))
	}
	cr.router.Use(middlewares.GzipDecompressionMiddleware)
	cr.router.Use(middlewares.SignatureMiddleware(cr.cfg.Key, cr.logger))

	cr.router.Route("/", func(router chi.Router) {
		router.Route("/", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("text/html")).
				Get("/", mh.Render)
		})
		router.Route("/update", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("application/json")).
				Post("/", mh.UpdateModel)
		})
		router.Route("/updates", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("application/json")).
				Post("/", mh.UpdateModels)
		})
		router.Route("/update/{metricType}/{metricName}/{value}", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("application/text")).
				Post("/", mh.Update)
		})
		router.Route("/value", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("application/json")).
				Post("/", mh.RetrieveModel)
		})
		router.Route("/value/{metricType}/{metricName}", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("application/text")).
				Get("/", mh.Retrieve)
		})
		router.Route("/ping", func(router chi.Router) {
			router.With(middlewares.ContentMiddleware("application/text")).
				Get("/", hh.Ping)
		})
	})
}

func (cr *CustomRouter) GetRouter() *chi.Mux {
	return cr.router
}
