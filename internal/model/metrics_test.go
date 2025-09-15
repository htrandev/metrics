package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMetricType(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected MetricType
	}{
		{
			name:     "gauge",
			value:    "gauge",
			expected: TypeGauge,
		},
		{
			name:     "counter",
			value:    "counter",
			expected: TypeCounter,
		},
		{
			name:     "unknown",
			value:    "unknown",
			expected: TypeUnknown,
		},
		{
			name:     "invalid",
			value:    "none",
			expected: TypeUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mt := ParseMetricType(tc.value)
			require.EqualValues(t, tc.expected, mt)
		})
	}
}

func TestSetValue(t *testing.T) {
	testCases := []struct {
		name           string
		metric         Metric
		value          string
		expectedMetric Metric
		wantErr        bool
	}{
		{
			name: "valid gauge",
			metric: Metric{
				Name: "gauge",
				Type: TypeGauge,
			},
			value: "0.1",
			expectedMetric: Metric{
				Name:  "gauge",
				Type:  TypeGauge,
				Value: MetricValue{Gauge: 0.1},
			},
			wantErr: false,
		},
		{
			name: "valid counter",
			metric: Metric{
				Name: "counter",
				Type: TypeCounter,
			},
			value: "1",
			expectedMetric: Metric{
				Name:  "counter",
				Type:  TypeCounter,
				Value: MetricValue{Counter: 1},
			},
			wantErr: false,
		},
		{
			name: "invalid counter value",
			metric: Metric{
				Name: "counter",
				Type: TypeCounter,
			},
			value:   "0.1",
			wantErr: true,
		},
		{
			name: "invalid gauge value",
			metric: Metric{
				Name: "gauge",
				Type: TypeGauge,
			},
			value:   "none",
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
			require.Equal(t, tc.expectedMetric, tc.metric)
		})
	}
}
