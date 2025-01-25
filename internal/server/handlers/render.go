package handlers

import (
	"net/http"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

// Render handles the rendering of the metrics page.
//
// Parameters:
//   - response: The HTTP response writer.
//   - request: The HTTP request.
//
// Behavior:
//   - Fetches all metrics from the repository.
//   - Reads the "index.html" template using the embedded reader.
//   - Renders the template with the metrics data.
//   - Returns HTTP 500 status code if any errors occur during template loading or rendering.
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
