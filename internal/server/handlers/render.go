package handlers

import (
	"net/http"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

func (mh *MetricHandler) Render(response http.ResponseWriter, request *http.Request) {
	page := struct {
		Metrics map[domain.MetricName]*db.Metric
	}{
		Metrics: mh.repo.GetAll(request.Context()),
	}
	tmpl, err := mh.embedReader.Read("index.html")
	if err != nil {
		mh.logger.Error().Err(err).Msg("Could not load template")
		http.Error(response, "Failed to load template: "+err.Error(), http.StatusInternalServerError)

		return
	}

	if err := tmpl.Execute(response, page); err != nil {
		mh.logger.Error().Err(err).Msg("Could not render template")
		http.Error(response, "Failed to render page: "+err.Error(), http.StatusInternalServerError)
	}
}
