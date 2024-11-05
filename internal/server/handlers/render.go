package handlers

import (
	"context"
	"net/http"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/web"
)

func (mh *MetricHandler) Render(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), mh.timeout)
	defer cancel()

	page := struct {
		Metrics map[domain.MetricName]db.MtrMetric
	}{
		Metrics: mh.repo.GetAll(ctx),
	}

	reader := web.NewEmbedReader()
	tmpl, err := reader.Read("index.html")
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
