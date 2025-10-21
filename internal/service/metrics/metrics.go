package metrics

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
)

const defaultInterval = 300 * time.Second

var defaultOpts = &ServiseOptions{
	StoreInterval: defaultInterval,
	Logger:        &zap.Logger{},
}

type Storage interface {
	Get(ctx context.Context, name string) (model.Metric, error)
	GetAll(ctx context.Context) ([]model.Metric, error)
	Store(ctx context.Context, metric *model.Metric) error
	Ping(ctx context.Context) error
}

type FileStorage interface {
	Flush(context.Context, []model.Metric) error
}

type ServiseOptions struct {
	StoreInterval time.Duration
	Logger        *zap.Logger

	Flusher FileStorage
	Storage Storage
}

type MetricsService struct {
	opts *ServiseOptions
}

func NewService(opts *ServiseOptions) *MetricsService {
	if opts.StoreInterval == 0 {
		opts.StoreInterval = defaultInterval
	}
	return &MetricsService{
		opts: opts,
	}
}

func (s *MetricsService) Ping(ctx context.Context) error {
	return s.opts.Storage.Ping(ctx)
}

func (s *MetricsService) Get(ctx context.Context, name string) (model.Metric, error) {
	m, err := s.opts.Storage.Get(ctx, name)
	if err != nil {
		return model.Metric{}, fmt.Errorf("get metric: %w", err)
	}
	return m, nil
}

func (s *MetricsService) GetAll(ctx context.Context) ([]model.Metric, error) {
	m, err := s.opts.Storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all metrics: %w", err)
	}
	return m, nil
}

func (s *MetricsService) Store(ctx context.Context, m *model.Metric) error {
	if m == nil {
		return nil
	}

	if err := s.opts.Storage.Store(ctx, m); err != nil {
		return fmt.Errorf("get all metrics: %w", err)
	}
	return nil
}

func (s *MetricsService) Run(ctx context.Context) {
	ticker := time.NewTicker(s.opts.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := s.opts.Storage.GetAll(ctx)
			if err != nil {
				s.opts.Logger.Error("get all metrics to flush", zap.Error(err), zap.String("scope", "Run"))
				continue
			}
			if err := s.opts.Flusher.Flush(ctx, metrics); err != nil {
				s.opts.Logger.Error("flush metrics", zap.Error(err), zap.String("scope", "Run"))
				continue
			}
		}
	}
}
