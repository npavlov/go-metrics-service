package dbmanager

import (
	"database/sql"

	"github.com/pressly/goose"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// DBManager manages the database connection and its lifecycle.
type DBManager struct {
	DB               *sql.DB
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

func (m *DBManager) Connect() *DBManager {
	sqlDB, err := sql.Open("pgx", m.connectionString)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database")

		return m
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database")

		return m
	}

	m.DB = sqlDB
	m.IsConnected = true

	return m
}

func (m *DBManager) ApplyMigrations() *DBManager {
	if !m.IsConnected {
		return m
	}

	// Run migrations
	if err := goose.Up(m.DB, "migrations"); err != nil {
		log.Error().Err(err).Msg("Failed to up migrations")
	}

	log.Info().Msg("Migrations completed")

	return m
}

// Close closes the underlying sql.DB connection.
func (m *DBManager) Close() {
	if m.DB == nil {
		return
	}

	if err := m.DB.Close(); err != nil {
		m.Log.Error().Err(err).Msg("Failed to close database connection")
	} else {
		m.Log.Info().Msg("Database connection closed")
	}
}
