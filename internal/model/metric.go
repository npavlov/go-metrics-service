package model

import (
	"github.com/npavlov/go-metrics-service/internal/domain"
)

type Metric struct {
	ID      domain.MetricName   `json:"id"`
	MType   domain.MetricType   `json:"type"`
	MSource domain.MetricSource `json:"-"`
	Delta   *int64              `json:"delta,omitempty"`
	Value   *float64            `json:"value,omitempty"`
}

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
