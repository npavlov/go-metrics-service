package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
)

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
		if oldMetric, found := oldMetrics[metric.ID]; found {
			oldMetric.SetValue(metric.Delta, metric.Value)
			oldMetrics[metric.ID] = oldMetric
		} else {
			oldMetrics[metric.ID] = *metric
		}
	}

	// Prepare the newMetrics slice with the updated oldMetrics
	newMetrics := make([]model.Metric, 0, len(oldMetrics))

	// Add all metrics from oldMetrics to newMetrics
	for _, oldMetric := range oldMetrics {
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