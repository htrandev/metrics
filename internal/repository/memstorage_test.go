package repository

import (
	"testing"

	models "github.com/htrandev/metrics/internal/model"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	emptyMemstorage := NewMemStorageRepository()

	testCases := []struct {
		name    string
		storage *MemStorage
		req     *models.Metric
		wantErr bool
	}{
		{
			name: "valid gauge",
			req: &models.Metric{
				Name:  "gauge",
				Type:  models.TypeGauge,
				Value: models.MetricValue{Gauge: 0.1},
			},
			storage: emptyMemstorage,
			wantErr: false,
		},
		{
			name: "valid counter",
			req: &models.Metric{
				Name:  "counter",
				Type:  models.TypeGauge,
				Value: models.MetricValue{Counter: 1},
			},
			storage: emptyMemstorage,
			wantErr: false,
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
				Type:  models.TypeGauge,
				Value: models.MetricValue{Gauge: 0.2},
			},
			wantErr: false,
		},
		{
			name:    "filled mem storage counter",
			storage: filledMemStorage(t),
			req: &models.Metric{
				Name:  "counter",
				Type:  models.TypeCounter,
				Value: models.MetricValue{Counter: 2},
			},
			wantErr: false,
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
		})
	}
}

func filledMemStorage(t *testing.T) *MemStorage {
	memstorage := NewMemStorageRepository()
	if err := memstorage.Store(&models.Metric{
		Name:  "gauge",
		Type:  models.TypeGauge,
		Value: models.MetricValue{Gauge: 0.1},
	}); err != nil {
		t.Fatalf("store gauge: %v", err)
	}

	if err := memstorage.Store(&models.Metric{
		Name:  "counter",
		Type:  models.TypeCounter,
		Value: models.MetricValue{Counter: 1},
	}); err != nil {
		t.Fatalf("store counter: %v", err)
	}
	return memstorage
}
