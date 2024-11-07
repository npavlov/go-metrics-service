package db

import (
	"strconv"

	"github.com/npavlov/go-metrics-service/internal/domain"
)

type Metric struct {
	CounterMetric
	GaugeMetric
	MtrMetric
}

func NewMetric(id domain.MetricName, mType domain.MetricType, delta *int64, value *float64) *Metric {
	return &Metric{
		CounterMetric: CounterMetric{
			Delta:    delta,
			MetricID: "",
		},
		GaugeMetric: GaugeMetric{
			Value:    value,
			MetricID: "",
		},
		MtrMetric: MtrMetric{
			ID:    id,
			MType: mType,
		},
	}
}

func (m *Metric) FromFields(id domain.MetricName, mType domain.MetricType, delta *int64, value *float64) {
	m.MtrMetric.ID = id
	m.MType = mType
	m.Delta = delta
	m.Value = value
}

// SetValue - the method that allows to encapsulate value set logic for different types.
func (m *Metric) SetValue(delta *int64, value *float64) {
	if m.MType == domain.Gauge {
		m.Delta = nil
		m.Value = value

		return
	}

	if m.MType == domain.Counter {
		m.Value = nil

		if m.Delta != nil && delta != nil {
			newDelta := *m.Delta + *delta
			m.Delta = &newDelta

			return
		}

		if delta != nil {
			m.Delta = delta

			return
		}
	}
}

// GetValue - the method that gets value for dedicated type.
func (m *Metric) GetValue() string {
	if m.MType == domain.Gauge {
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	}

	if m.MType == domain.Counter {
		return strconv.FormatInt(*m.Delta, 10)
	}

	return ""
}
