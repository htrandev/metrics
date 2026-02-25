package metrics

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/htrandev/metrics/internal/model"
)

var (
	errStore              = errors.New("store error")
	errStoreMany          = errors.New("store many error")
	errStoreManyWithRetry = errors.New("store many with retry error")

	errGet    = errors.New("get error")
	errGetAll = errors.New("getAll error")

	errPing = errors.New("ping error")
)

var _ model.Storager = (*mockStorage)(nil)

type mockStorage struct {
	getErr                bool
	getAllErr             bool
	storeErr              bool
	storeManyErr          bool
	storeManyWithRetryErr bool
	pingErr               bool

	gauge  bool
	filled bool
}

func (m *mockStorage) Ping(_ context.Context) error {
	if m.pingErr {
		return errPing
	}
	return nil
}

func (m *mockStorage) Close() error { return nil }

func (m *mockStorage) Set(_ context.Context, _ *model.MetricDto) error {
	return nil
}

func (m *mockStorage) Get(_ context.Context, _ string) (model.MetricDto, error) {
	if m.getErr {
		return model.MetricDto{}, errGet
	}
	metric := model.MetricDto{Name: "test"}
	if m.gauge {
		metric.Value.Type = model.TypeGauge
		metric.Value.Gauge = 0.1
		return metric, nil
	}

	metric.Value.Type = model.TypeCounter
	metric.Value.Counter = 1

	return metric, nil
}

func (m *mockStorage) GetAll(_ context.Context) ([]model.MetricDto, error) {
	if m.getAllErr {
		return nil, errGetAll
	}
	if m.filled {
		return filledStorage(), nil
	}
	return []model.MetricDto{}, nil
}

func (m *mockStorage) Store(_ context.Context, _ *model.MetricDto) error {
	if m.storeErr {
		return errStore
	}
	return nil
}

func (m *mockStorage) StoreMany(_ context.Context, _ []model.MetricDto) error {
	if m.storeManyErr {
		return errStoreMany
	}
	return nil
}

func (m *mockStorage) StoreManyWithRetry(_ context.Context, _ []model.MetricDto) error {
	if m.storeManyWithRetryErr {
		return errStoreManyWithRetry
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
		expectedMetric model.MetricDto
		expectedError  error
	}{
		{
			name:       "valid gauge",
			storage:    &mockStorage{gauge: true},
			metricName: "test",
			wantErr:    false,
			expectedMetric: model.MetricDto{
				Name:  "test",
				Value: model.MetricValue{Type: model.TypeGauge, Gauge: 0.1},
			},
		},
		{
			name:       "valid counter",
			storage:    &mockStorage{gauge: false},
			metricName: "test",
			wantErr:    false,
			expectedMetric: model.MetricDto{
				Name:  "test",
				Value: model.MetricValue{Type: model.TypeCounter, Counter: 1},
			},
		},
		{
			name:           "invalid",
			storage:        &mockStorage{getErr: true},
			metricName:     "test",
			wantErr:        true,
			expectedMetric: model.MetricDto{},
			expectedError:  errGet,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(&ServiсeOptions{Storage: tc.storage})

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
		expectedMetric []model.MetricDto
		expectedError  error
	}{
		{
			name:           "valid empty",
			storage:        &mockStorage{filled: false},
			wantErr:        false,
			expectedMetric: []model.MetricDto{},
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
			s := NewService(&ServiсeOptions{Storage: tc.storage})

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
		metric        *model.MetricDto
		wantErr       bool
		expectedError error
	}{
		{
			name:    "valid gauge",
			storage: &mockStorage{},
			metric:  &model.MetricDto{},
			wantErr: false,
		},
		{
			name:          "invalid",
			storage:       &mockStorage{storeErr: true},
			metric:        &model.MetricDto{},
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
			s := NewService(&ServiсeOptions{Storage: tc.storage})

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

func TestStoreMany(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name          string
		storage       *mockStorage
		metrics       []model.MetricDto
		wantErr       bool
		expectedError error
	}{
		{
			name:    "valid",
			storage: &mockStorage{},
			metrics: []model.MetricDto{
				model.Gauge("gauge", 0.1),
				model.Counter("counter", 1),
			},
			wantErr: false,
		},
		{
			name:    "invalid",
			storage: &mockStorage{storeManyErr: true},
			metrics: []model.MetricDto{
				model.Gauge("gauge", 0.1),
				model.Counter("counter", 1),
			},
			wantErr:       true,
			expectedError: errStoreMany,
		},
		{
			name:    "nil request",
			storage: &mockStorage{},
			metrics: nil,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(&ServiсeOptions{Storage: tc.storage})

			err := s.StoreMany(ctx, tc.metrics)
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestStoreManyWithRetry(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name          string
		storage       *mockStorage
		metrics       []model.MetricDto
		wantErr       bool
		expectedError error
	}{
		{
			name:    "valid",
			storage: &mockStorage{},
			metrics: []model.MetricDto{
				model.Gauge("gauge", 0.1),
				model.Counter("counter", 1),
			},
			wantErr: false,
		},
		{
			name:    "invalid",
			storage: &mockStorage{storeManyWithRetryErr: true},
			metrics: []model.MetricDto{
				model.Gauge("gauge", 0.1),
				model.Counter("counter", 1),
			},
			wantErr:       true,
			expectedError: errStoreManyWithRetry,
		},
		{
			name:    "nil request",
			storage: &mockStorage{},
			metrics: nil,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(&ServiсeOptions{Storage: tc.storage})

			err := s.StoreManyWithRetry(ctx, tc.metrics)
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPing(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		storage       *mockStorage
		wantErr       bool
		expectedError error
	}{
		{
			name:    "valid",
			storage: &mockStorage{},
			wantErr: false,
		},
		{
			name:          "invalid",
			storage:       &mockStorage{pingErr: true},
			wantErr:       true,
			expectedError: errPing,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(&ServiсeOptions{Storage: tc.storage})
			err := s.Ping(ctx)
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}

func filledStorage() []model.MetricDto {
	metrics := []model.MetricDto{
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
