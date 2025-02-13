package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

// Retrieve handles HTTP requests to retrieve a specific metric by its name.
// It retrieves the metric from the repository and writes its value to the response.
func (mh *MetricHandler) Retrieve(response http.ResponseWriter, request *http.Request) {
	metricName := domain.MetricName(chi.URLParam(request, "metricName"))

	metricModel, found := mh.repo.Get(request.Context(), metricName)
	if !found {
		log.Error().Msgf("Retrieve: Failed to retrieve model from memory %s", metricName)
		http.Error(response, "Failed to retrieve model from memory", http.StatusNotFound)

		return
	}

	_, _ = response.Write([]byte(metricModel.GetValue()))
	response.WriteHeader(http.StatusOK)
}

// RetrieveModel handles HTTP requests to retrieve a metric using JSON input.
// It decodes the input, retrieves the metric from the repository, and sends it as JSON in the response.
func (mh *MetricHandler) RetrieveModel(response http.ResponseWriter, request *http.Request) {
	// Decode the incoming JSON request into the Metric struct
	var metric *db.Metric
	if err := mh.json.NewDecoder(request.Body).Decode(&metric); err != nil {
		mh.logger.Error().Err(err).Msg("Invalid JSON input")
		http.Error(response, "Invalid JSON input", http.StatusNotFound)

		return
	}

	// Prepare the updated metric to be returned
	responseMetric, found := mh.repo.Get(request.Context(), metric.ID)
	if !found {
		mh.logger.Error().Msgf("Failed to retrieve model from memory %s", metric.ID)
		http.Error(response, "Failed to retrieve model from memory", http.StatusNotFound)

		return
	}

	response.WriteHeader(http.StatusOK)
	err := mh.json.NewEncoder(response).Encode(responseMetric)
	if err != nil {
		mh.logger.Error().Err(err).Msg("Failed to encode response JSON")
		http.Error(response, "Failed to process response", http.StatusInternalServerError)
	}
}
