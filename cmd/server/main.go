package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
	"github.com/npavlov/go-metrics-service/internal/utils"
	"github.com/pkg/errors"
	"github.com/pressly/goose"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log := logger.NewLogger().SetLogLevel(zerolog.DebugLevel).Get()

	err := godotenv.Load("server.env")
	if err != nil {
		log.Error().Err(err).Msg("Error loading server.env file")
	}

	cfg := config.NewConfigBuilder(log).
		FromEnv().
		FromFlags().Build()

	log.Info().Interface("config", cfg).Msg("Configuration loaded")

	ctx, cancel := utils.WithSignalCancel(context.Background(), log)

	db, err := getDB(cfg.Database)
	if err != nil {
		log.Warn().Err(err).Msg("Error connecting to database")
	}
	defer closeDB(db, log)

	memStorage := storage.NewMemStorage(log).WithBackup(ctx, cfg)
	// set repo to dbStorage if we are using database
	dbStorage := storage.NewDBStorage(db, log)

	stMonitor := storage.NewAutoSwitchRepo(memStorage, dbStorage, cfg.HealthCheckDur, log)
	stMonitor.StartMonitoring(ctx)

	mHandlers := handlers.NewMetricsHandler(stMonitor, log)
	hHandlers := handlers.NewHealthHandler(dbStorage, log)
	var cRouter router.Router = router.NewCustomRouter(log)
	cRouter.SetRouter(mHandlers, hHandlers)

	log.Info().
		Str("server_address", cfg.Address).
		Msg("Server started")

	//nolint:exhaustruct
	server := &http.Server{
		Addr:         cfg.Address,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		Handler:      cRouter.GetRouter(),
	}

	go func() {
		// Wait for the context to be done (i.e., signal received)
		<-ctx.Done()

		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Error shutting down server")
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Err(err).Msg("Error starting server")
		cancel()
	}

	log.Info().Msg("Server shut down")
}

func getDB(connectionString string) (*sql.DB, error) {
	if connectionString == "" {
		return nil, errors.New("no connection string provided")
	}

	sqlDB, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to database")
	}

	// Run migrations
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Error().Err(err).Msg("Failed to up migrations")
	}

	log.Info().Msg("Migrations completed")

	return sqlDB, nil
}

// closeDB retrieves and closes the underlying sql.DB connection.
func closeDB(db *sql.DB, log *zerolog.Logger) {
	if db == nil {
		return
	}

	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close database connection")
	} else {
		log.Info().Msg("Database connection closed")
	}
}
