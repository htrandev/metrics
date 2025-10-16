package memstorage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/htrandev/metrics/internal/model"
)

func TestStore(t *testing.T) {
	emptyMemstorage := NewRepository()

	testCases := []struct {
		name          string
		storage       *MemStorage
		req           *model.Metric
		wantErr       bool
		expectedValue model.Metric
	}{
		{
			name: "valid gauge",
			req: &model.Metric{
				Name:  "gauge",
				Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1},
			},
			storage: emptyMemstorage,
			wantErr: false,
			expectedValue: model.Metric{
				Name:  "gauge",
				Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1},
			},
		},
		{
			name: "valid counter",
			req: &model.Metric{
				Name:  "counter",
				Value: model.MetricValue{Type: model.TypeCounter, Counter: 1},
			},
			storage: emptyMemstorage,
			wantErr: false,
			expectedValue: model.Metric{
				Name:  "counter",
				Value: model.MetricValue{Type: model.TypeCounter, Counter: 1},
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
			req: &model.Metric{
				Name:  "gauge",
				Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.2},
			},
			wantErr: false,
			expectedValue: model.Metric{
				Name:  "gauge",
				Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.2},
			},
		},
		{
			name:    "filled mem storage counter",
			storage: filledMemStorage(t),
			req: &model.Metric{
				Name:  "counter",
				Value: model.MetricValue{Type: model.TypeCounter, Counter: 2},
			},
			wantErr: false,
			expectedValue: model.Metric{
				Name:  "counter",
				Value: model.MetricValue{Type: model.TypeCounter, Counter: 3},
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

			var actValue model.Metric
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
	emptyMemstorage := NewRepository()

	testCases := []struct {
		name           string
		storage        *MemStorage
		metricName     string
		wantErr        bool
		expectedMetric model.Metric
	}{
		{
			name:           "valid gauge",
			storage:        filledMemStorage(t),
			metricName:     "gauge",
			wantErr:        false,
			expectedMetric: model.Gauge("gauge", 0.1),
		},
		{
			name:           "valid counter",
			storage:        filledMemStorage(t),
			metricName:     "counter",
			wantErr:        false,
			expectedMetric: model.Counter("counter", 1),
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
	emptyMemstorage := NewRepository()

	testCases := []struct {
		name           string
		storage        *MemStorage
		wantErr        bool
		expectedResult []model.Metric
	}{
		{
			name:           "valid empty storage",
			storage:        emptyMemstorage,
			wantErr:        false,
			expectedResult: []model.Metric{},
		},
		{
			name:    "valid filled storage",
			storage: filledMemStorage(t),
			wantErr: false,
			expectedResult: []model.Metric{
				{Name: "counter", Value: model.MetricValue{
					Type:    model.TypeCounter,
					Counter: 1,
				}},
				{Name: "gauge", Value: model.MetricValue{
					Type:  model.TypeGauge,
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

func TestSet(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		storage        *MemStorage
		req            *model.Metric
		expetedMetrics []model.Metric
	}{
		{
			name:    "valid empty gauge",
			storage: NewRepository(),
			req: &model.Metric{
				Name:  "gauge",
				Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.2},
			},
			expetedMetrics: []model.Metric{{
				Name:  "gauge",
				Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.2},
			}},
		},
		{
			name:    "valid empty counter",
			storage: NewRepository(),
			req: &model.Metric{
				Name:  "counter",
				Value: model.MetricValue{Type: model.TypeCounter, Counter: 2},
			},
			expetedMetrics: []model.Metric{{
				Name:  "counter",
				Value: model.MetricValue{Type: model.TypeCounter, Counter: 2},
			}},
		},
		{
			name:           "valid nil request",
			storage:        NewRepository(),
			req:            nil,
			expetedMetrics: []model.Metric{},
		},
		{
			name:    "valid filled gauge",
			storage: filledMemStorage(t),
			req: &model.Metric{
				Name:  "gauge",
				Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.2},
			},
			expetedMetrics: []model.Metric{
				{
					Name:  "counter",
					Value: model.MetricValue{Type: model.TypeCounter, Counter: 1},
				},
				{
					Name:  "gauge",
					Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1},
				},
			},
		},
		{
			name:    "valid filled counter",
			storage: filledMemStorage(t),
			req: &model.Metric{
				Name:  "counter",
				Value: model.MetricValue{Type: model.TypeCounter, Counter: 2},
			},
			expetedMetrics: []model.Metric{
				{
					Name:  "counter",
					Value: model.MetricValue{Type: model.TypeCounter, Counter: 1},
				},
				{
					Name:  "gauge",
					Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.storage.Set(ctx, tc.req)
			require.NoError(t, err)

			metrics, err := tc.storage.GetAll(ctx)
			require.NoError(t, err)

			require.Equal(t, tc.expetedMetrics, metrics)
		})
	}
}

func filledMemStorage(t *testing.T) *MemStorage {
	t.Helper()
	ctx := context.Background()

	memstorage := NewRepository()
	if err := memstorage.Store(ctx, &model.Metric{
		Name:  "gauge",
		Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1},
	}); err != nil {
		t.Fatalf("store gauge: %v", err)
	}

	if err := memstorage.Store(ctx, &model.Metric{
		Name:  "counter",
		Value: model.MetricValue{Type: model.TypeCounter, Counter: 1},
	}); err != nil {
		t.Fatalf("store counter: %v", err)
	}
	return memstorage
}
