package db

import (
	"strconv"

	"github.com/npavlov/go-metrics-service/internal/domain"
)

// SetValue - the method that allows to encapsulate value set logic for different types.
func (m *MtrMetric) SetValue(delta *int64, value *float64) {
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
func (m *MtrMetric) GetValue() string {
	if m.MType == domain.Gauge {
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	}

	if m.MType == domain.Counter {
		return strconv.FormatInt(*m.Delta, 10)
	}

	return ""
}

// ToUpsertMetricParams converts to params.
func (m *MtrMetric) ToUpsertMetricParams() UpsertMetricParams {
	return UpsertMetricParams{
		ID:    m.ID,
		MType: m.MType,
		Delta: m.Delta,
		Value: m.Value,
	}
}

// ToInsertMetricParams converts to params.
func (m *MtrMetric) ToInsertMetricParams() InsertMetricParams {
	return InsertMetricParams{
		ID:    m.ID,
		MType: m.MType,
		Delta: m.Delta,
		Value: m.Value,
	}
}

// ToUpdateMetricParams  converts to params.
func (m *MtrMetric) ToUpdateMetricParams() UpdateMetricParams {
	return UpdateMetricParams{
		ID:    m.ID,
		MType: m.MType,
		Delta: m.Delta,
		Value: m.Value,
	}
}
