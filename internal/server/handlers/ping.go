package handlers

import (
	"net/http"
)

func (mh *MetricHandler) Ping(response http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if err := mh.universalDB.Repo.Ping(ctx); err != nil {
		mh.logger.Error().Err(err).Msg("No connection to database")
		http.Error(response, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
	}
}
