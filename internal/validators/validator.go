package validators

import (
	"fmt"
	"io"
	"strconv"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

// MValidator - the interface to describe validators for metrics.
type MValidator interface {
	FromVars(mName domain.MetricName, mType domain.MetricType, val string) (*db.Metric, error)
	FromBody(body io.ReadCloser) (*db.Metric, error)
	ManyFromBody(body io.ReadCloser) ([]*db.Metric, error)
	ValidateStructure(metric *db.Metric) error
}

// MValidatorImpl - the implementation structure for validations.
type MValidatorImpl struct {
	validate *validator.Validate
	json     jsoniter.API
}

// NewMetricsValidator - the builder function for MValidatorImpl.
func NewMetricsValidator() *MValidatorImpl {
	return &MValidatorImpl{
		validate: validator.New(validator.WithRequiredStructEnabled()),
		json:     jsoniter.ConfigCompatibleWithStandardLibrary,
	}
}

// FromVars - the function that parses metric structure from map object.
func (v *MValidatorImpl) FromVars(mName domain.MetricName, mType domain.MetricType, val string) (*db.Metric, error) {
	metric := &db.Metric{
		MtrMetric: db.MtrMetric{
			ID:    "",
			MType: "",
		},
		GaugeMetric: db.GaugeMetric{
			MetricID: "",
			Value:    nil,
		},
		CounterMetric: db.CounterMetric{
			MetricID: "",
			Delta:    nil,
		},
	}
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
func (v *MValidatorImpl) FromBody(body io.ReadCloser) (*db.Metric, error) {
	metric := &db.Metric{
		MtrMetric: db.MtrMetric{
			ID:    "",
			MType: "",
		},
		GaugeMetric: db.GaugeMetric{
			MetricID: "",
			Value:    nil,
		},
		CounterMetric: db.CounterMetric{
			MetricID: "",
			Delta:    nil,
		},
	}

	err := v.json.NewDecoder(body).Decode(metric)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse metric json")
	}

	if metric.MType != domain.Counter && metric.MType != domain.Gauge {
		return nil, fmt.Errorf("failed to validate metric type: %s", metric.MType)
	}

	if metric.MType == domain.Counter {
		metric.Value = nil
	}

	if metric.MType == domain.Gauge {
		metric.Delta = nil
	}

	return metric, nil
}

// ManyFromBody - the function that parses many metric structures from reader.
func (v *MValidatorImpl) ManyFromBody(body io.ReadCloser) ([]*db.Metric, error) {
	var metrics []*db.Metric

	err := v.json.NewDecoder(body).Decode(&metrics)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse metric json")
	}

	for _, metric := range metrics {
		if metric.MType != domain.Counter && metric.MType != domain.Gauge {
			return nil, fmt.Errorf("failed to validate metric type: %s", metric.MType)
		}

		if metric.MType == domain.Counter {
			metric.Value = nil
		}

		if metric.MType == domain.Gauge {
			metric.Delta = nil
		}
	}

	return metrics, nil
}

func (v *MValidatorImpl) ValidateStructure(metric *db.Metric) error {
	return v.validate.Struct(metric)
}
