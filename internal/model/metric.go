package model

import (
	"github.com/npavlov/go-metrics-service/internal/domain"
	"strconv"
)

type Metric struct {
	ID      domain.MetricName
	MType   domain.MetricType
	MSource domain.MetricSource
	Counter *int64
	Value   *float64
}

func (m *Metric) SetValue(counter *int64, value *float64) {
	if m.MType == domain.Gauge {
		m.Counter = nil
		m.Value = value
		return
	}

	if m.MType == domain.Counter {
		m.Value = nil

		if m.Counter != nil && counter != nil {
			newDelta := *m.Counter + *counter
			m.Counter = &newDelta
			return
		}

		if counter != nil {
			m.Counter = counter
			return
		}
	}
}

func (m *Metric) GetValue() (string, bool) {
	if m.MType == domain.Gauge && m.Value != nil {
		return strconv.FormatFloat(*m.Value, 'f', -1, 64), true
	}

	if m.MType == domain.Counter && m.Counter != nil {
		return strconv.FormatInt(*m.Counter, 10), true
	}

	return "", false
}
