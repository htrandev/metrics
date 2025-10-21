package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/htrandev/metrics/internal/model"
)

type PostgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	if err := r.db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Close(ctx context.Context) error {
	return r.db.Close()
}

func (r *PostgresRepository) Get(ctx context.Context, name string) (model.Metric, error) {
	return model.Metric{}, nil
}
func (r *PostgresRepository) GetAll(ctx context.Context) ([]model.Metric, error)    { return nil, nil }
func (r *PostgresRepository) Store(ctx context.Context, metric *model.Metric) error { return nil }
func (r *PostgresRepository) Set(ctx context.Context, metric *model.Metric) error   { return nil }
