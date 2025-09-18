package repository

import (
	"testing"

	"github.com/stretchr/testify/require"

	models "github.com/htrandev/metrics/internal/model"
)

func TestStore(t *testing.T) {
	emptyMemstorage := NewMemStorageRepository()

	testCases := []struct {
		name          string
		storage       *MemStorage
		req           *models.Metric
		wantErr       bool
		expectedValue models.Metric
	}{
		{
			name: "valid gauge",
			req: &models.Metric{
				Name:  "gauge",
				Value: models.MetricValue{Type: models.TypeGauge, Gauge: 0.1},
			},
			storage: emptyMemstorage,
			wantErr: false,
			expectedValue: models.Metric{
				Name:  "gauge",
				Value: models.MetricValue{Type: models.TypeGauge, Gauge: 0.1},
			},
		},
		{
			name: "valid counter",
			req: &models.Metric{
				Name:  "counter",
				Value: models.MetricValue{Type: models.TypeCounter, Counter: 1},
			},
			storage: emptyMemstorage,
			wantErr: false,
			expectedValue: models.Metric{
				Name:  "counter",
				Value: models.MetricValue{Type: models.TypeCounter, Counter: 1},
			},
		},
		{
			name:    "nil request",
			storage: emptyMemstorage,
			req:     nil,
			wantErr: false,
		},
		{
			name:    "filled mem storage gauge",
			storage: filledMemStorage(t),
			req: &models.Metric{
				Name:  "gauge",
				Value: models.MetricValue{Type: models.TypeGauge, Gauge: 0.2},
			},
			wantErr: false,
			expectedValue: models.Metric{
				Name:  "gauge",
				Value: models.MetricValue{Type: models.TypeGauge, Gauge: 0.2},
			},
		},
		{
			name:    "filled mem storage counter",
			storage: filledMemStorage(t),
			req: &models.Metric{
				Name:  "counter",
				Value: models.MetricValue{Type: models.TypeCounter, Counter: 2},
			},
			wantErr: false,
			expectedValue: models.Metric{
				Name:  "counter",
				Value: models.MetricValue{Type: models.TypeCounter, Counter: 3},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.storage.Store(tc.req)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, tc.storage.metrics[tc.expectedValue.Name])
		})
	}
}

func filledMemStorage(t *testing.T) *MemStorage {
	t.Helper()

	memstorage := NewMemStorageRepository()
	if err := memstorage.Store(&models.Metric{
		Name:  "gauge",
		Value: models.MetricValue{Type: models.TypeGauge, Gauge: 0.1},
	}); err != nil {
		t.Fatalf("store gauge: %v", err)
	}

	if err := memstorage.Store(&models.Metric{
		Name:  "counter",
		Value: models.MetricValue{Type: models.TypeCounter, Counter: 1},
	}); err != nil {
		t.Fatalf("store counter: %v", err)
	}
	return memstorage
}
