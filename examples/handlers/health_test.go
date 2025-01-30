package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/npavlov/go-metrics-service/internal/server/dbmanager"
	"github.com/npavlov/go-metrics-service/internal/server/handlers"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
	"github.com/pashagolub/pgxmock/v4"
)

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
	defer result.Body.Close()

	// Print status code
	fmt.Println(result.StatusCode)

	// Output:
	// 200
}
