package handlers

import (
	"net/http"

	"github.com/npavlov/go-metrics-service/internal/server/repository"
	"github.com/rs/zerolog"
)

type HealthHandler struct {
	logger   *zerolog.Logger
	database *repository.DBRepository
}

// NewHealthHandler - constructor for HealthHandler.
func NewHealthHandler(database *repository.DBRepository, l *zerolog.Logger) *HealthHandler {
	return &HealthHandler{
		logger:   l,
		database: database,
	}
}

func (mh *HealthHandler) Ping(response http.ResponseWriter, _ *http.Request) {
	if err := mh.database.Ping(); err != nil {
		mh.logger.Error().Err(err).Msg("No connection to database")
		http.Error(response, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
	}

	response.WriteHeader(http.StatusOK)
}
