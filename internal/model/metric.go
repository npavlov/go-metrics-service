package model

import (
	"github.com/npavlov/go-metrics-service/internal/domain"
	"strconv"
)

type Metric struct {
	ID      domain.MetricName   `json:"id"`
	MType   domain.MetricType   `json:"type"`
	MSource domain.MetricSource `json:"-"`
	Delta   *int64              `json:"delta,omitempty"`
	Value   *float64            `json:"value,omitempty"`
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
