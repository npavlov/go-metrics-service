package testutils

import (
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/dbmanager"
	"github.com/npavlov/go-metrics-service/internal/server/storage"
)

// SetupDBManager creates a DBManager with a pgxmock pool and returns the DBManager and mock pool.
//
//nolint:ireturn
func SetupDBManager(t *testing.T) (*dbmanager.DBManager, pgxmock.PgxPoolIface, *zerolog.Logger) {
	t.Helper()

	mockDB, err := pgxmock.NewPool()
	require.NoError(t, err)

	log := GetTLogger()
	dbStorage := dbmanager.NewDBManager("mock connection string", log)
	dbStorage.DB = mockDB
	dbStorage.IsConnected = true

	return dbStorage, mockDB, log
}

// SetupDBStorage creates a DBStorage with a pgxmock pool and returns the DBStorate and mock pool.
//
//nolint:ireturn
func SetupDBStorage(t *testing.T) (*storage.DBStorage, pgxmock.PgxPoolIface) {
	t.Helper()

	mockDB, err := pgxmock.NewPool()
	require.NoError(t, err)

	log := GetTLogger()
	dbStorage := storage.NewDBStorage(mockDB, log)

	return dbStorage, mockDB
}
