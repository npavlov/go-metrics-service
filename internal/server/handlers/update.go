package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
)

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
	err = json.NewEncoder(response).Encode(metric)
	if err != nil {
		mh.logger.Error().Err(err).Msg("Failed to encode response JSON")
		http.Error(response, "Failed to process response", http.StatusInternalServerError)
	}
}

func (mh *MetricHandler) UpdateModels(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), mh.timeout)
	defer cancel()

	// Parse and validate metrics from the request body
	metrics, err := mh.validator.ManyFromBody(request.Body)
	if err != nil {
		mh.logger.Error().Err(err).Msg("error validating structure")
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	// Collect metric IDs for database retrieval
	metricIDs := make([]domain.MetricName, len(metrics))
	for i, metric := range metrics {
		metricIDs[i] = metric.ID
	}

	// Fetch old metrics for updating existing ones
	oldMetrics, err := mh.repo.GetMany(ctx, metricIDs)
	if err != nil {
		mh.logger.Error().Err(err).Msg("error getting old metrics")
		http.Error(response, err.Error(), http.StatusInternalServerError)

		return
	}

	// Prepare new metrics by updating existing ones or creating new entries
	for _, metric := range metrics {
		if oldMetric, found := (*oldMetrics)[metric.ID]; found {
			oldMetric.SetValue(metric.Delta, metric.Value)
			(*oldMetrics)[metric.ID] = oldMetric
		} else {
			(*oldMetrics)[metric.ID] = *metric
		}
	}

	// Prepare the newMetrics slice with the updated oldMetrics
	newMetrics := make([]model.Metric, 0, len(*oldMetrics))

	// Add all metrics from oldMetrics to newMetrics
	for _, oldMetric := range *oldMetrics {
		newMetrics = append(newMetrics, oldMetric)
	}

	// Update all metrics in the repository
	if err = mh.repo.UpdateMany(ctx, &newMetrics); err != nil {
		mh.logger.Error().Err(err).Msg("error updating metrics")
		http.Error(response, "Failed to update metrics", http.StatusInternalServerError)

		return
	}

	// Send the response with updated metrics
	response.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(response).Encode(newMetrics); err != nil {
		mh.logger.Error().Err(err).Msg("Failed to encode response JSON")
		http.Error(response, "Failed to process response", http.StatusInternalServerError)
	}
}

func (mh *MetricHandler) updateAndReturn(request *http.Request, newMetric *model.Metric) (*model.Metric, error) {
	ctx, cancel := context.WithTimeout(request.Context(), mh.timeout)
	defer cancel()

	existingMetric, found := mh.repo.Get(ctx, newMetric.ID)

	if found {
		existingMetric.SetValue(newMetric.Delta, newMetric.Value)

		err := mh.repo.Update(ctx, existingMetric)
		if err != nil {
			mh.logger.Error().Err(err).Msg("error updating existingMetric")

			return nil, errors.Wrap(err, "error updating existingMetric")
		}

		return existingMetric, nil
	}

	err := mh.repo.Create(ctx, newMetric)
	if err != nil {
		mh.logger.Error().Err(err).Msg("error creating Metric")

		return nil, errors.Wrap(err, "error creating new Metric")
	}

	return newMetric, nil
}
