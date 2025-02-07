package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/domain"
)

func TestMetricType_Scan(t *testing.T) {
	t.Parallel()

	var m domain.MetricType
	err := m.Scan("gauge")
	require.NoError(t, err)
	assert.Equal(t, domain.Gauge, m)

	err = m.Scan(123)
	assert.ErrorIs(t, err, domain.ErrInvalidStr)
}

func TestMetricType_Value(t *testing.T) {
	t.Parallel()

	m := domain.Gauge
	val, err := m.Value()
	require.NoError(t, err)
	assert.Equal(t, "gauge", val)
}

func TestMetricName_String(t *testing.T) {
	t.Parallel()

	m := domain.Alloc
	assert.Equal(t, "Alloc", m.String())
}
