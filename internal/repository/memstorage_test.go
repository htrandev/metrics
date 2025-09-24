package repository

import (
	"context"
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
			err := tc.storage.Store(context.Background(), tc.req)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			var actValue models.Metric
			if tc.expectedValue.Name != "" {
				actValue, err = tc.storage.Get(context.Background(), tc.expectedValue.Name)
				require.NoError(t, err)
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, actValue)
		})
	}
}

func TestGet(t *testing.T) {
	emptyMemstorage := NewMemStorageRepository()

	testCases := []struct {
		name           string
		storage        *MemStorage
		metricName     string
		wantErr        bool
		expectedMetric models.Metric
	}{
		{
			name:           "valid gauge",
			storage:        filledMemStorage(t),
			metricName:     "gauge",
			wantErr:        false,
			expectedMetric: models.Gauge("gauge", 0.1),
		},
		{
			name:           "valid counter",
			storage:        filledMemStorage(t),
			metricName:     "counter",
			wantErr:        false,
			expectedMetric: models.Counter("counter", 1),
		},
		{
			name:       "empty storage",
			storage:    emptyMemstorage,
			metricName: "test",
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := tc.storage.Get(context.Background(), tc.metricName)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.EqualValues(t, tc.expectedMetric, m)
		})
	}
}

func TestGetAll(t *testing.T) {
	emptyMemstorage := NewMemStorageRepository()

	testCases := []struct {
		name           string
		storage        *MemStorage
		wantErr        bool
		expectedResult []models.Metric
	}{
		{
			name:           "valid empty storage",
			storage:        emptyMemstorage,
			wantErr:        false,
			expectedResult: []models.Metric{},
		},
		{
			name:    "valid filled storage",
			storage: filledMemStorage(t),
			wantErr: false,
			expectedResult: []models.Metric{
				{Name: "counter", Value: models.MetricValue{
					Type:    models.TypeCounter,
					Counter: 1,
				}},
				{Name: "gauge", Value: models.MetricValue{
					Type:  models.TypeGauge,
					Gauge: 0.1,
				}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := tc.storage.GetAll(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.EqualValues(t, tc.expectedResult, m)
		})
	}
}

func filledMemStorage(t *testing.T) *MemStorage {
	t.Helper()
	ctx := context.Background()

	memstorage := NewMemStorageRepository()
	if err := memstorage.Store(ctx, &models.Metric{
		Name:  "gauge",
		Value: models.MetricValue{Type: models.TypeGauge, Gauge: 0.1},
	}); err != nil {
		t.Fatalf("store gauge: %v", err)
	}

	if err := memstorage.Store(ctx, &models.Metric{
		Name:  "counter",
		Value: models.MetricValue{Type: models.TypeCounter, Counter: 1},
	}); err != nil {
		t.Fatalf("store counter: %v", err)
	}
	return memstorage
}
