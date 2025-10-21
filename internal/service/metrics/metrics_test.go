package metrics

import (
	"context"
	"errors"
	"testing"

	"github.com/htrandev/metrics/internal/model"
	"github.com/stretchr/testify/require"
)

var (
	errStore  = errors.New("store error")
	errGet    = errors.New("get error")
	errGetAll = errors.New("getAll error")
)

var _ model.Storager = (*mockStorage)(nil)

type mockStorage struct {
	getErr    bool
	getAllErr bool
	storeErr  bool

	gauge  bool
	filled bool
}

func (m *mockStorage) Ping(_ context.Context) error {
	return nil
}

func (m *mockStorage) Set(_ context.Context, _ *model.Metric) error {
	return nil
}

func (m *mockStorage) Get(_ context.Context, _ string) (model.Metric, error) {
	if m.getErr {
		return model.Metric{}, errGet
	}
	metric := model.Metric{Name: "test"}
	if m.gauge {
		metric.Value.Type = model.TypeGauge
		metric.Value.Gauge = 0.1
		return metric, nil
	}

	metric.Value.Type = model.TypeCounter
	metric.Value.Counter = 1

	return metric, nil
}

func (m *mockStorage) GetAll(_ context.Context) ([]model.Metric, error) {
	if m.getAllErr {
		return nil, errGetAll
	}
	if m.filled {
		return filledStorage(), nil
	}
	return []model.Metric{}, nil
}

func (m *mockStorage) Store(_ context.Context, _ *model.Metric) error {
	if m.storeErr {
		return errStore
	}
	return nil
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name           string
		storage        *mockStorage
		metricName     string
		wantErr        bool
		expectedMetric model.Metric
		expectedError  error
	}{
		{
			name:       "valid gauge",
			storage:    &mockStorage{gauge: true},
			metricName: "test",
			wantErr:    false,
			expectedMetric: model.Metric{
				Name:  "test",
				Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1},
			},
		},
		{
			name:       "valid counter",
			storage:    &mockStorage{gauge: false},
			metricName: "test",
			wantErr:    false,
			expectedMetric: model.Metric{
				Name:  "test",
				Value: model.MetricValue{Type: model.TypeCounter, Counter: 1},
			},
		},
		{
			name:           "invalid",
			storage:        &mockStorage{getErr: true},
			metricName:     "test",
			wantErr:        true,
			expectedMetric: model.Metric{},
			expectedError:  errGet,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(&ServiseOptions{Storage: tc.storage})

			m, err := s.Get(ctx, tc.metricName)
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedMetric, m)

		})
	}
}
func TestGetAll(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name           string
		storage        *mockStorage
		wantErr        bool
		expectedMetric []model.Metric
		expectedError  error
	}{
		{
			name:           "valid empty",
			storage:        &mockStorage{filled: false},
			wantErr:        false,
			expectedMetric: []model.Metric{},
		},
		{
			name:           "valid filled",
			storage:        &mockStorage{filled: true},
			wantErr:        false,
			expectedMetric: filledStorage(),
		},
		{
			name:           "invalid",
			storage:        &mockStorage{getAllErr: true},
			wantErr:        true,
			expectedMetric: nil,
			expectedError:  errGetAll,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(&ServiseOptions{Storage: tc.storage})

			m, err := s.GetAll(ctx)
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedMetric, m)

		})
	}
}

func TestStore(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name          string
		storage       *mockStorage
		metric        *model.Metric
		wantErr       bool
		expectedError error
	}{
		{
			name:    "valid gauge",
			storage: &mockStorage{},
			metric:  &model.Metric{},
			wantErr: false,
		},
		{
			name:          "invalid",
			storage:       &mockStorage{storeErr: true},
			metric:        &model.Metric{},
			wantErr:       true,
			expectedError: errStore,
		},
		{
			name:    "nil request",
			storage: &mockStorage{},
			metric:  nil,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(&ServiseOptions{Storage: tc.storage})

			err := s.Store(ctx, tc.metric)
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)

		})
	}
}

func filledStorage() []model.Metric {
	metrics := []model.Metric{
		{
			Name: "gauge",
			Value: model.MetricValue{
				Type:  model.TypeGauge,
				Gauge: 0.1,
			},
		},
		{
			Name: "counter",
			Value: model.MetricValue{
				Type:    model.TypeCounter,
				Counter: 1,
			},
		},
	}
	return metrics
}
