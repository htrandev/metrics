package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/stretchr/testify/require"

	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupTesting(t *testing.T) *PostgresRepository {
	t.Helper()

	const (
		dsn            = "postgres://user:password@localhost:5432/practicum"
		migrationsPath = "../../../migrations"
		truncateQuery  = `TRUNCATE TABLE metrics`
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)

	provider, err := goose.NewProvider(database.DialectPostgres, db, migrations.Embed)
	require.NoError(t, err)

	_, err = provider.Up(ctx)
	require.NoError(t, err)

	r := New(db)

	t.Cleanup(func() {
		r.Truncate(context.Background())
		r.Close()
	})
	return r
}

func TestPing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	r := setupTesting(t)

	err := r.Ping(ctx)
	require.NoError(t, err)
}

func TestSet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	setGauge := "set gauge"
	setCounter := "set counter"

	r := setupTesting(t)

	testCases := []struct {
		name           string
		metric         *model.Metric
		expectedMetric model.Metric
	}{
		{
			name: "valid gauge first",
			metric: func() *model.Metric {
				m := model.Gauge(setGauge, 0.1)
				return &m
			}(),
			expectedMetric: model.Gauge(setGauge, 0.1),
		},
		{
			name: "valid gauge second",
			metric: func() *model.Metric {
				m := model.Gauge(setGauge, 0.2)
				return &m
			}(),
			expectedMetric: model.Gauge(setGauge, 0.2),
		},
		{
			name: "valid counter first",
			metric: func() *model.Metric {
				m := model.Counter(setCounter, 1)
				return &m
			}(),
			expectedMetric: model.Counter(setCounter, 1),
		},
		{
			name: "valid counter second",
			metric: func() *model.Metric {
				m := model.Counter(setCounter, 2)
				return &m
			}(),
			expectedMetric: model.Counter(setCounter, 2),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := r.Set(ctx, tc.metric)
			require.NoError(t, err)

			m, err := r.Get(ctx, tc.metric.Name)
			require.NoError(t, err)
			require.Equal(t, tc.expectedMetric, m)
		})
	}
}

func TestStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	storeGauge := "store gauge"
	storeCounter := "store counter"

	r := setupTesting(t)

	testCases := []struct {
		name           string
		metric         *model.Metric
		expectedMetric model.Metric
	}{
		{
			name: "valid gauge first",
			metric: func() *model.Metric {
				m := model.Gauge(storeGauge, 0.1)
				return &m
			}(),
			expectedMetric: model.Gauge(storeGauge, 0.1),
		},
		{
			name: "valid gauge second",
			metric: func() *model.Metric {
				m := model.Gauge(storeGauge, 0.2)
				return &m
			}(),
			expectedMetric: model.Gauge(storeGauge, 0.2),
		},
		{
			name: "valid counter first",
			metric: func() *model.Metric {
				m := model.Counter(storeCounter, 1)
				return &m
			}(),
			expectedMetric: model.Counter(storeCounter, 1),
		},
		{
			name: "valid counter second",
			metric: func() *model.Metric {
				m := model.Counter(storeCounter, 2)
				return &m
			}(),
			expectedMetric: model.Counter(storeCounter, 3),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := r.Store(ctx, tc.metric)
			require.NoError(t, err)

			m, err := r.Get(ctx, tc.metric.Name)
			require.NoError(t, err)
			require.Equal(t, tc.expectedMetric, m)
		})
	}
}

func TestGet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	getGauge := "get gauge"
	getCounter := "get counter"

	r := setupTesting(t)

	testCases := []struct {
		name           string
		metric         *model.Metric
		expectedMetric model.Metric
	}{
		{
			name: "valid gauge first",
			metric: func() *model.Metric {
				m := model.Gauge(getGauge, 0.1)
				return &m
			}(),
			expectedMetric: model.Gauge(getGauge, 0.1),
		},
		{
			name: "valid gauge second",
			metric: func() *model.Metric {
				m := model.Gauge(getGauge, 0.2)
				return &m
			}(),
			expectedMetric: model.Gauge(getGauge, 0.2),
		},
		{
			name: "valid counter first",
			metric: func() *model.Metric {
				m := model.Counter(getCounter, 1)
				return &m
			}(),
			expectedMetric: model.Counter(getCounter, 1),
		},
		{
			name: "valid counter second",
			metric: func() *model.Metric {
				m := model.Counter(getCounter, 2)
				return &m
			}(),
			expectedMetric: model.Counter(getCounter, 3),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := r.Store(ctx, tc.metric)
			require.NoError(t, err)

			m, err := r.Get(ctx, tc.expectedMetric.Name)
			require.NoError(t, err)
			require.Equal(t, tc.expectedMetric, m)
		})
	}
}

func TestGetAll(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	getAllGauge := "get all gauge"
	getAllCounter := "get all counter"

	r := setupTesting(t)

	testCases := []struct {
		name            string
		metrics         []*model.Metric
		expectedMetrics []model.Metric
	}{
		{
			name:            "valid empty",
			metrics:         nil,
			expectedMetrics: []model.Metric{},
		},
		{
			name: "valid not empty",
			metrics: func() []*model.Metric {
				mg1 := model.Gauge(getAllGauge, 0.1)
				mg2 := model.Gauge(getAllGauge, 0.2)
				mc1 := model.Counter(getAllCounter, 1)
				mc2 := model.Counter(getAllCounter, 2)
				metrics := []*model.Metric{&mg1, &mg2, &mc1, &mc2}
				return metrics
			}(),
			expectedMetrics: []model.Metric{
				model.Gauge(getAllGauge, 0.2),
				model.Counter(getAllCounter, 3),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.metrics) > 0 {
				for _, metric := range tc.metrics {
					err := r.Store(ctx, metric)
					require.NoError(t, err)
				}
			}

			m, err := r.GetAll(ctx)
			require.NoError(t, err)
			require.Equal(t, tc.expectedMetrics, m)
		})
	}
}
