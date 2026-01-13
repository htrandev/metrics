package metrics

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
)

// Storage предоставляет интерфейс для работы с хранилищем.
type Storage interface {
	Get(ctx context.Context, name string) (model.Metric, error)
	GetAll(ctx context.Context) ([]model.Metric, error)

	Store(ctx context.Context, metric *model.Metric) error
	StoreMany(ctx context.Context, metric []model.Metric) error
	StoreManyWithRetry(ctx context.Context, metric []model.Metric) error

	Ping(ctx context.Context) error
}

// ServiceOptions определяет параметры сервиса.
type ServiсeOptions struct {
	Logger *zap.Logger

	Storage Storage
}

// MetricsService определяет сервис для работы с метриками
type MetricsService struct {
	opts *ServiсeOptions
}

// NewService возвращает новый экземпляр сервиса.
func NewService(opts *ServiсeOptions) *MetricsService {
	return &MetricsService{
		opts: opts,
	}
}

// Ping проверяет доступность хранилища.
func (s *MetricsService) Ping(ctx context.Context) error {
	return s.opts.Storage.Ping(ctx)
}

// Get возвращает метрику по имени.
func (s *MetricsService) Get(ctx context.Context, name string) (model.Metric, error) {
	m, err := s.opts.Storage.Get(ctx, name)
	if err != nil {
		return model.Metric{}, fmt.Errorf("get metric: %w", err)
	}
	return m, nil
}

// GetAll возвращает все метрики.
func (s *MetricsService) GetAll(ctx context.Context) ([]model.Metric, error) {
	m, err := s.opts.Storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all metrics: %w", err)
	}
	return m, nil
}

// Store сохраняет одну метрику.
func (s *MetricsService) Store(ctx context.Context, m *model.Metric) error {
	if m == nil {
		return nil
	}

	if err := s.opts.Storage.Store(ctx, m); err != nil {
		return fmt.Errorf("store metric: %w", err)
	}
	return nil
}

// StoreMany сохраняет батч метрик.
func (s *MetricsService) StoreMany(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	if err := s.opts.Storage.StoreMany(ctx, metrics); err != nil {
		return fmt.Errorf("store many metrics: %w", err)
	}
	return nil
}

// StoreManyWithRetry сохраняет батч с повторными попытками.
func (s *MetricsService) StoreManyWithRetry(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	if err := s.opts.Storage.StoreManyWithRetry(ctx, metrics); err != nil {
		return fmt.Errorf("store many with retry metrics: %w", err)
	}

	return nil
}
