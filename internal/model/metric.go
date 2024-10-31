package model

import (
	"strconv"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"gorm.io/gorm"
)

type Metric struct {
	// Standard fields for gorm
	gorm.Model
	ID      domain.MetricName   `json:"id"      gorm:"primaryKey;type:varchar(255);column:id;not null"        validate:"required"`
	MType   domain.MetricType   `json:"type"    gorm:"column:type;not null"        validate:"required,oneof=counter gauge"`
	MSource domain.MetricSource `json:"-" gorm:"-"`
	Delta   *int64              `json:"delta,omitempty" gorm:"column:delta;type:bigint"`
	Value   *float64            `json:"value,omitempty" gorm:"column:value;type:double precision"`
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
