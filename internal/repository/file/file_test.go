package file

import (
	"context"
	"os"
	"testing"

	"github.com/htrandev/metrics/internal/model"
	"github.com/stretchr/testify/require"
)

func TestFlush(t *testing.T) {
	ctx := context.Background()

	fileName := "temp_test.log"
	tempFile, err := os.CreateTemp("", fileName)
	require.NoError(t, err)

	defer os.Remove(fileName)

	testCases := []struct {
		name    string
		metrics []model.Metric
	}{
		{
			name: "valid",
			metrics: []model.Metric{
				{Name: "gauge", Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1}},
				{Name: "counter", Value: model.MetricValue{Type: model.TypeCounter, Gauge: 1}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &FileRepository{file: tempFile}

			err := r.Flush(ctx, tc.metrics)
			require.NoError(t, err)
		})
	}
}
