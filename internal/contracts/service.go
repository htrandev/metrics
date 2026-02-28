package contracts

import (
	"context"

	"github.com/htrandev/metrics/internal/model"
)

// Service предоставляет интерфейс взаимодействия с сервисом для работы с метриками.
type Service interface {
	Get(ctx context.Context, name string) (model.MetricDto, error)
	GetAll(ctx context.Context) ([]model.MetricDto, error)

	Store(ctx context.Context, metric *model.MetricDto) error
	StoreMany(ctx context.Context, metric []model.MetricDto) error
	StoreManyWithRetry(ctx context.Context, metric []model.MetricDto) error

	Ping(ctx context.Context) error
}
