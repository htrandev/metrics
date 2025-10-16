package metrics

import (
	"context"
	"fmt"

	"github.com/htrandev/metrics/internal/model"
)

type Storage interface {
	Get(ctx context.Context, name string) (model.Metric, error)
	GetAll(ctx context.Context) ([]model.Metric, error)
	Store(ctx context.Context, metric *model.Metric) error
}

type MetricsService struct {
	storage Storage
}

func NewService(s Storage) *MetricsService {
	return &MetricsService{
		storage: s,
	}
}

func (s *MetricsService) Get(ctx context.Context, name string) (model.Metric, error) {
	m, err := s.storage.Get(ctx, name)
	if err != nil {
		return model.Metric{}, fmt.Errorf("get metric: %w", err)
	}
	return m, nil
}

func (s *MetricsService) GetAll(ctx context.Context) ([]model.Metric, error) {
	m, err := s.storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all metrics: %w", err)
	}
	return m, nil
}

func (s *MetricsService) Store(ctx context.Context, m *model.Metric) error {
	if m == nil {
		return nil
	}

	if err := s.storage.Store(ctx, m); err != nil {
		return fmt.Errorf("get all metrics: %w", err)
	}
	return nil
}
