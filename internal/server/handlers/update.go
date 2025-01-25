package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

// Update handles HTTP requests to update a specific metric by its type, name, and value.
// It validates the input, updates the metric, and sends an appropriate HTTP response.
func (mh *MetricHandler) Update(response http.ResponseWriter, request *http.Request) {
	metricType := domain.MetricType(chi.URLParam(request, "metricType"))
	metricName := domain.MetricName(chi.URLParam(request, "metricName"))
	metricValue := chi.URLParam(request, "value")

	newMetric, err := mh.validator.FromVars(metricName, metricType, metricValue)
	if err != nil {
		mh.logger.Error().Err(err).Msg("error validating structure")
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	_, err = mh.updateAndReturn(request, newMetric)
	if err != nil {
		mh.logger.Error().Err(err).Msg("error updating metric")
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	response.WriteHeader(http.StatusOK)
}

// UpdateModel handles HTTP requests to update a metric using JSON input.
// It validates the input, updates the metric, and sends the updated metric as JSON in the response.
func (mh *MetricHandler) UpdateModel(response http.ResponseWriter, request *http.Request) {
	newMetric, err := mh.validator.FromBody(request.Body)
	if err != nil {
		mh.logger.Error().Err(err).Msg("error validating structure")
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	metric, err := mh.updateAndReturn(request, newMetric)
	if err != nil {
		mh.logger.Error().Err(err).Msg("error updating metric")
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	response.WriteHeader(http.StatusOK)
	err = mh.json.NewEncoder(response).Encode(metric)
	if err != nil {
		mh.logger.Error().Err(err).Msg("Failed to encode response JSON")
		http.Error(response, "Failed to process response", http.StatusInternalServerError)
	}
}

// updateAndReturn is a helper function to update or create a metric in the repository.
// It retrieves an existing metric or creates a new one, updates its value, and returns the updated metric.
func (mh *MetricHandler) updateAndReturn(request *http.Request, newMetric *db.Metric) (*db.Metric, error) {
	existingMetric, found := mh.repo.Get(request.Context(), newMetric.ID)

	if found {
		existingMetric.SetValue(newMetric.Delta, newMetric.Value)

		err := mh.repo.Update(request.Context(), existingMetric)
		if err != nil {
			mh.logger.Error().Err(err).Msg("error updating existingMetric")

			return nil, errors.Wrap(err, "error updating existingMetric")
		}

		return existingMetric, nil
	}

	err := mh.repo.Create(request.Context(), newMetric)
	if err != nil {
		mh.logger.Error().Err(err).Msg("error creating Metric")

		return nil, errors.Wrap(err, "error creating new Metric")
	}

	return newMetric, nil
}
