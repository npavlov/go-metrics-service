package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"

	"github.com/npavlov/go-metrics-service/internal/server/handlers"
)

var errPingError = errors.New("ping error")

// TestHealthHandlerPing_Successful tests the Ping handler when the database is connected and ping is successful.
func TestHealthHandlerPing_Successful(t *testing.T) {
	t.Parallel()

	dbManager, mock, log := testutils.SetupDBManager(t)
	defer mock.Close()

	handler := handlers.NewHealthHandler(dbManager, log)
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

	dbManager, _, log := testutils.SetupDBManager(t)
	dbManager.IsConnected = false
	defer func() { dbManager.IsConnected = true }() // Reset for other tests

	handler := handlers.NewHealthHandler(dbManager, log)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	handler.Ping(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

// TestHealthHandlerPing_PingFails tests the Ping handler when the database ping fails.
func TestHealthHandlerPing_PingFails(t *testing.T) {
	t.Parallel()

	dbManager, mock, log := testutils.SetupDBManager(t)
	defer mock.Close()

	handler := handlers.NewHealthHandler(dbManager, log)
	mock.ExpectPing().WillReturnError(errPingError)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	handler.Ping(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
