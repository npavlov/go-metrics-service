package handlers_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/dbmanager"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
)

var errPingError = errors.New("ping error")

// setupDBStorage creates a DBManager with a pgxmock pool and returns the DBManager and mock pool.
func setupDBStorage(t *testing.T) (*dbmanager.DBManager, pgxmock.PgxPoolIface, *zerolog.Logger) {
	t.Helper()

	mockDB, err := pgxmock.NewPool()
	require.NoError(t, err)

	log := logger.NewLogger().Get()
	dbStorage := dbmanager.NewDBManager("mock connection string", log)
	dbStorage.DB = mockDB
	dbStorage.IsConnected = true

	return dbStorage, mockDB, log
}

// TestHealthHandlerPing_Successful tests the Ping handler when the database is connected and ping is successful.
func TestHealthHandlerPing_Successful(t *testing.T) {
	t.Parallel()

	dbStorage, mock, log := setupDBStorage(t)
	defer mock.Close()

	handler := handlers.NewHealthHandler(dbStorage, log)
	mock.ExpectPing().WillReturnError(nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	handler.Ping(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestHealthHandlerPing_DBNotConnected tests the Ping handler when the database is not connected.
func TestHealthHandlerPing_DBNotConnected(t *testing.T) {
	t.Parallel()

	dbStorage, _, log := setupDBStorage(t)
	dbStorage.IsConnected = false
	defer func() { dbStorage.IsConnected = true }() // Reset for other tests

	handler := handlers.NewHealthHandler(dbStorage, log)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	handler.Ping(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

// TestHealthHandlerPing_PingFails tests the Ping handler when the database ping fails.
func TestHealthHandlerPing_PingFails(t *testing.T) {
	t.Parallel()

	dbStorage, mock, log := setupDBStorage(t)
	defer mock.Close()

	handler := handlers.NewHealthHandler(dbStorage, log)
	mock.ExpectPing().WillReturnError(errPingError)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	handler.Ping(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func ExampleHealthHandler_Ping() {
	mockDB, _ := pgxmock.NewPool()
	mockDB.ExpectPing().WillReturnError(nil)

	log := testutils.GetTLogger()
	dbStorage := dbmanager.NewDBManager("mock connection string", log)
	dbStorage.DB = mockDB
	dbStorage.IsConnected = true

	handler := handlers.NewHealthHandler(dbStorage, log)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	handler.Ping(resp, req)

	result := resp.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(result.Body)

	// Print status code
	fmt.Println(result.StatusCode)

	// Output:
	// 200
}
