package validators

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/pkg/errors"
)

// Validator - the interface to describe validator for metrics.
type Validator interface {
	FromVars(mName domain.MetricName, mType domain.MetricType, val string) (*model.Metric, error)
	FromBody(body io.ReadCloser) (*model.Metric, error)
}

// ValidatorImpl - the implementation structure for validations.
type ValidatorImpl struct {
	validate *validator.Validate
}

// NewMetricsValidator - the builder function for ValidatorImpl.
func NewMetricsValidator() *ValidatorImpl {
	return &ValidatorImpl{
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

// FromVars - the function that parses metric structure from map object.
func (v *ValidatorImpl) FromVars(mName domain.MetricName, mType domain.MetricType, val string) (*model.Metric, error) {
	metric := &model.Metric{}
	// Retrieving variables
	if len(mName) == 0 {
		return nil, errors.New("failed to retrieve metricID path param")
	}

	if len(mType) == 0 {
		return nil, errors.New("failed to retrieve mType path param")
	}
	// Assigning values
	metric.ID = mName

	if len(val) == 0 {
		return nil, errors.New("failed to retrieve value path param")
	}

	if mType == domain.Counter {
		metric.MType = domain.Counter

		delta, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse value")
		}

		metric.Delta = &delta
		metric.Value = nil
	}

	if mType == domain.Gauge {
		metric.MType = domain.Gauge

		value, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse value")
		}

		metric.Value = &value
		metric.Delta = nil
	}

	// Validate structure
	err := v.validate.Struct(metric)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate metric")
	}

	return metric, nil
}

// FromBody - the function that parses metric structure from reader.
func (v *ValidatorImpl) FromBody(body io.ReadCloser) (*model.Metric, error) {
	metric := &model.Metric{}

	err := json.NewDecoder(body).Decode(metric)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse metric json")
	}

	if metric.MType != domain.Counter && metric.MType != domain.Gauge {
		return nil, errors.Wrap(err, "failed to validate metric type: %s")
	}

	if metric.MType == domain.Counter {
		metric.Value = nil
	}

	if metric.MType == domain.Gauge {
		metric.Delta = nil
	}

	return metric, nil
}
