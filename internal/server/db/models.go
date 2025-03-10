// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql/driver"
	"fmt"

	domain "github.com/npavlov/go-metrics-service/internal/domain"
)

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

func (e *MetricType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = MetricType(s)
	case string:
		*e = MetricType(s)
	default:
		return fmt.Errorf("unsupported scan type for MetricType: %T", src)
	}
	return nil
}

type NullMetricType struct {
	MetricType MetricType
	Valid      bool // Valid is true if MetricType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullMetricType) Scan(value interface{}) error {
	if value == nil {
		ns.MetricType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.MetricType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullMetricType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.MetricType), nil
}

type CounterMetric struct {
	MetricID domain.MetricName `db:"metric_id" json:"-"`
	Delta    *int64            `db:"delta" json:"delta"`
}

type GaugeMetric struct {
	MetricID domain.MetricName `db:"metric_id" json:"-"`
	Value    *float64          `db:"value" json:"value"`
}

type MtrMetric struct {
	ID    domain.MetricName `db:"id" json:"id" validate:"required"`
	MType domain.MetricType `db:"type" json:"type" validate:"required,oneof=counter gauge"`
}
