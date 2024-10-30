package handlers

import (
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

	_, err = mh.updateAndReturn(newMetric)
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

	metric, err := mh.updateAndReturn(newMetric)
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

func (mh *MetricHandler) updateAndReturn(newMetric *model.Metric) (*model.Metric, error) {
	existingMetric, found := mh.universalDB.Storage.Get(newMetric.ID)

	if found {
		existingMetric.SetValue(newMetric.Delta, newMetric.Value)

		err := mh.universalDB.Storage.Update(existingMetric)
		if err != nil {
			mh.logger.Error().Err(err).Msg("error updating existingMetric")

			return nil, errors.Wrap(err, "error updating existingMetric")
		}

		return existingMetric, nil
	}

	err := mh.universalDB.Storage.Create(newMetric)
	if err != nil {
		mh.logger.Error().Err(err).Msg("error creating Metric")

		return nil, errors.Wrap(err, "error creating new Metric")
	}

	return newMetric, nil
}
