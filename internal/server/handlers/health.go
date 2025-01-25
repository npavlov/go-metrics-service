package handlers

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/server/dbmanager"
)

// @Title Health API
// @Description Service to get DB state
// @Version 1.0

// @Contact.email test@test.com

// @BasePath /
// @Host localhost:8080

// @SecurityDefinitions.apikey ApiKeyAuth
// @In header
// @Name authorization

// @Tag.name Health
// @Tag.description "Handlers that provide information about current state"

type HealthHandler struct {
	logger   *zerolog.Logger
	database *dbmanager.DBManager
}

// NewHealthHandler - constructor for HealthHandler.
func NewHealthHandler(database *dbmanager.DBManager, l *zerolog.Logger) *HealthHandler {
	return &HealthHandler{
		logger:   l,
		database: database,
	}
}

// Ping
// @Summary      Health check for the service
// @Description  Checks the health of the service by verifying the database connection.
// @Tags         Health
// @Produce      plain
// @Success      200  {string}  string  "Service is healthy"
// @Failure      500  {string}  string  "Failed to connect to database"
// @Router       /ping [get]
func (mh *HealthHandler) Ping(response http.ResponseWriter, req *http.Request) {
	if !mh.database.IsConnected {
		mh.logger.Info().Msg("Database is not connected, can't ping")
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	if err := mh.database.DB.Ping(req.Context()); err != nil {
		mh.logger.Error().Err(err).Msg("No connection to database")
		http.Error(response, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)

		return
	}

	response.WriteHeader(http.StatusOK)
}
