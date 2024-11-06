package dbmanager

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type PgxPool interface {
	Exec(ctx context.Context, query string, options ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, options ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, options ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
	Ping(ctx context.Context) error
	Close()
}

// DBManager manages the database connection and its lifecycle.
type DBManager struct {
	DB               PgxPool
	Log              *zerolog.Logger
	connectionString string
	IsConnected      bool
}

// NewDBManager creates a new DBManager and opens a database connection.
func NewDBManager(connectionString string, log *zerolog.Logger) *DBManager {
	return &DBManager{
		connectionString: connectionString,
		Log:              log,
		IsConnected:      false,
		DB:               nil,
	}
}

// Connect initializes a new pgx connection pool.
func (m *DBManager) Connect(ctx context.Context) *DBManager {
	config, err := pgxpool.ParseConfig(m.connectionString)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse database configuration")

		return m
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database")

		return m
	}

	if err := pool.Ping(ctx); err != nil {
		return m
	}

	m.DB = pool
	m.IsConnected = true

	return m
}

// ApplyMigrations applies migrations using goose.
func (m *DBManager) ApplyMigrations() *DBManager {
	if !m.IsConnected {
		return m
	}

	sqlDB, err := sql.Open("pgx", m.connectionString)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database")

		return m
	}

	// Run migrations
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Error().Err(err).Msg("Failed to apply migrations")
	}

	log.Info().Msg("Migrations completed")

	return m
}

// Close closes the underlying pgxpool.Pool connection.
func (m *DBManager) Close() {
	if m.DB == nil {
		return
	}

	m.DB.Close()
	m.Log.Info().Msg("Database connection closed")
}
