package validators_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/validators"
)

func TestMValidatorImpl_FromVars(t *testing.T) {
	t.Parallel()

	validator := validators.NewMetricsValidator()

	tests := []struct {
		name    string
		mName   domain.MetricName
		mType   domain.MetricType
		val     string
		want    *model.Metric
		wantErr bool
	}{
		{
			name:  "Valid counter metric",
			mName: "test_counter",
			mType: domain.Counter,
			val:   "123",
			want: &model.Metric{
				ID:    "test_counter",
				MType: domain.Counter,
				Delta: int64Ptr(123),
			},
			wantErr: false,
		},
		{
			name:  "Valid gauge metric",
			mName: "test_gauge",
			mType: domain.Gauge,
			val:   "123.45",
			want: &model.Metric{
				ID:    "test_gauge",
				MType: domain.Gauge,
				Value: float64Ptr(123.45),
			},
			wantErr: false,
		},
		{
			name:    "Invalid metric name",
			mName:   "",
			mType:   domain.Counter,
			val:     "123",
			wantErr: true,
		},
		{
			name:    "Invalid metric type",
			mName:   "test",
			mType:   "",
			val:     "123",
			wantErr: true,
		},
		{
			name:    "Invalid counter value",
			mName:   "test_counter",
			mType:   domain.Counter,
			val:     "not-a-number",
			wantErr: true,
		},
		{
			name:    "Invalid gauge value",
			mName:   "test_gauge",
			mType:   domain.Gauge,
			val:     "not-a-float",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := validator.FromVars(tt.mName, tt.mType, tt.val)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMValidatorImpl_FromBody(t *testing.T) {
	t.Parallel()

	validator := validators.NewMetricsValidator()

	tests := []struct {
		name    string
		body    string
		want    *model.Metric
		wantErr bool
	}{
		{
			name: "Valid counter metric",
			body: `{"id":"test_counter","type":"counter","delta":123}`,
			want: &model.Metric{
				ID:    "test_counter",
				MType: domain.Counter,
				Delta: int64Ptr(123),
			},
			wantErr: false,
		},
		{
			name: "Valid gauge metric",
			body: `{"id":"test_gauge","type":"gauge","value":123.45}`,
			want: &model.Metric{
				ID:    "test_gauge",
				MType: domain.Gauge,
				Value: float64Ptr(123.45),
			},
			wantErr: false,
		},
		{
			name:    "Invalid JSON body",
			body:    `{"id": "test","type": "unknown","delta":123}`,
			wantErr: true,
		},
		{
			name:    "Invalid metric type",
			body:    `{"id":"test_metric","type":"invalid","delta":123}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body := io.NopCloser(bytes.NewBufferString(tt.body))
			got, err := validator.FromBody(body)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMValidatorImpl_ManyFromBody(t *testing.T) {
	t.Parallel()

	validator := validators.NewMetricsValidator()

	tests := []struct {
		name    string
		body    string
		want    []*model.Metric
		wantErr bool
	}{
		{
			name: "Valid metrics list",
			body: `[{"id":"metric1","type":"counter","delta":10},{"id":"metric2","type":"gauge","value":20.5}]`,
			want: []*model.Metric{
				{ID: "metric1", MType: domain.Counter, Delta: int64Ptr(10)},
				{ID: "metric2", MType: domain.Gauge, Value: float64Ptr(20.5)},
			},
			wantErr: false,
		},
		{
			name:    "Mixed valid and invalid metric types",
			body:    `[{"id":"metric1","type":"counter","delta":10},{"id":"metric2","type":"invalid","value":20.5}]`,
			wantErr: true,
		},
		{
			name:    "Invalid JSON structure",
			body:    `[{"id":"metric1","type":"counter","delta":10,]`,
			wantErr: true,
		},
		{
			name:    "Empty metrics list",
			body:    `[]`,
			want:    []*model.Metric{},
			wantErr: false,
		},
		{
			name: "Gauge metric without value field",
			body: `[{"id":"metric2","type":"gauge"}]`,
			want: []*model.Metric{
				{ID: "metric2", MType: domain.Gauge, Value: nil},
			},
			wantErr: false,
		},
		{
			name: "Counter metric without delta field",
			body: `[{"id":"metric1","type":"counter"}]`,
			want: []*model.Metric{
				{ID: "metric1", MType: domain.Counter, Delta: nil},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body := io.NopCloser(bytes.NewBufferString(tt.body))
			got, err := validator.ManyFromBody(body)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// Helper function to create float64 pointer.
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create int64 pointer.
func int64Ptr(i int64) *int64 {
	return &i
}
