package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/internal/repository"
)

type PostgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	if err := r.db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresRepository) Truncate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `TRUNCATE TABLE metrics`)
	if err != nil {
		return fmt.Errorf("repository/truncate: exec: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Get(ctx context.Context, name string) (model.Metric, error) {
	query := `SELECT type, gauge, counter
		FROM metrics
		WHERE name = $1
		LIMIT 1
	;`

	row := r.db.QueryRowContext(ctx, query, name)
	var (
		t       model.MetricType
		gauge   sql.NullFloat64
		counter sql.NullInt64
	)

	if err := row.Scan(&t, &gauge, &counter); err != nil {
		if err == sql.ErrNoRows {
			return model.Metric{}, repository.ErrNotFound
		}
		return model.Metric{}, fmt.Errorf("repository/get: scan: %w", err)
	}

	if row.Err() != nil {
		return model.Metric{}, fmt.Errorf("repository/get: row: %w", row.Err())
	}

	return buildMetric(name, t, gauge.Float64, counter.Int64), nil
}

func (r *PostgresRepository) GetAll(ctx context.Context) ([]model.Metric, error) {
	query := `SELECT name, type, gauge, counter
		FROM metrics
	;`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository/getAll: query context: %w", err)
	}

	metrics := make([]model.Metric, 0)
	for rows.Next() {
		var (
			name    string
			t       model.MetricType
			gauge   sql.NullFloat64
			counter sql.NullInt64
		)

		if err := rows.Scan(&name, &t, &gauge, &counter); err != nil {
			return nil, fmt.Errorf("repository/getAll: scan: %w", err)
		}

		m := buildMetric(name, t, gauge.Float64, counter.Int64)
		metrics = append(metrics, m)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("repository/getAll: rows.Err(): %w", err)
	}

	return metrics, nil
}

func (r *PostgresRepository) Store(ctx context.Context, metric *model.Metric) error {
	query := `INSERT INTO metrics (name, type, gauge, counter) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, type)
		DO UPDATE SET 
			gauge = $3, 
			counter = metrics.counter + $4
	;`
	_, err := r.db.ExecContext(ctx, query,
		metric.Name,
		metric.Value.Type,
		metric.Value.Gauge,
		metric.Value.Counter,
	)
	if err != nil {
		return fmt.Errorf("repository/store: exec query: %w", err)
	}
	return nil
}

func (r *PostgresRepository) StoreMany(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	query := `INSERT INTO metrics (name, type, gauge, counter) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, type)
		DO UPDATE SET 
			gauge = $3, 
			counter = metrics.counter + $4
	;`

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("repository/storeMany: begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("repository/storeMany: prepare query: %w", err)
	}

	for _, metric := range metrics {
		_, err := stmt.ExecContext(ctx, metric.Name,
			metric.Value.Type,
			metric.Value.Gauge,
			metric.Value.Counter)
		if err != nil {
			return fmt.Errorf("repository/storeMany: exec stmt: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) Set(ctx context.Context, metric *model.Metric) error {
	query := `INSERT INTO metrics (name, type, gauge, counter) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, type) 
		DO UPDATE SET 
			gauge = $3, 
			counter = $4
	;`
	_, err := r.db.ExecContext(ctx, query,
		metric.Name,
		metric.Value.Type,
		metric.Value.Gauge,
		metric.Value.Counter,
	)
	if err != nil {
		return fmt.Errorf("repository/set: exec query: %w", err)
	}
	return nil
}

func buildMetric(name string, t model.MetricType, gauge float64, counter int64) model.Metric {
	var m model.Metric

	switch t {
	case model.TypeGauge:
		return model.Gauge(name, gauge)
	case model.TypeCounter:
		return model.Counter(name, counter)
	}

	return m
}
