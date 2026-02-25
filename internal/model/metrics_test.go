package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetValue(t *testing.T) {
	testCases := []struct {
		name           string
		value          string
		metric         MetricDto
		wantErr        bool
		expectedMetric MetricDto
	}{
		{
			name:           "valid gauge",
			value:          "0.1",
			metric:         MetricDto{Name: "gauge", Value: MetricValue{Type: TypeGauge}},
			wantErr:        false,
			expectedMetric: MetricDto{Name: "gauge", Value: MetricValue{Type: TypeGauge, Gauge: 0.1}},
		},
		{
			name:           "valid counter",
			value:          "1",
			metric:         MetricDto{Name: "counter", Value: MetricValue{Type: TypeCounter}},
			wantErr:        false,
			expectedMetric: MetricDto{Name: "counter", Value: MetricValue{Type: TypeCounter, Counter: 1}},
		},
		{
			name:    "invalid gauge value",
			value:   "test",
			metric:  MetricDto{Name: "gauge", Value: MetricValue{Type: TypeGauge}},
			wantErr: true,
		},
		{
			name:    "invalid counter value",
			value:   "test",
			metric:  MetricDto{Name: "counter", Value: MetricValue{Type: TypeCounter}},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.metric.SetValue(tc.value)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.EqualValues(t, tc.expectedMetric, tc.metric)
		})
	}
}
