package memstorage

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/htrandev/metrics/internal/model"
)

type MemStorage struct {
	metrics map[string]model.Metric
}

func NewRepository() *MemStorage {
	metrics := make(map[string]model.Metric)
	return &MemStorage{
		metrics: metrics,
	}
}

// Set записывает значение метрики.
// Если метрика уже существует, то ничего не делает.
func (m *MemStorage) Set(Ctx context.Context, request *model.Metric) error {
	if request == nil {
		log.Println("repository: request is nil")
		return nil
	}

	if _, ok := m.metrics[request.Name]; !ok {
		m.metrics[request.Name] = *request
	}

	return nil
}

// Store записывает новое значение метрики.
// Если метрика существует, то обновляет ее значение.
func (m *MemStorage) Store(ctx context.Context, request *model.Metric) error {
	if request == nil {
		log.Println("repository: request is nil")
		return nil
	}
	metric, ok := m.metrics[request.Name]
	if !ok {
		m.metrics[request.Name] = *request
		return nil
	}

	switch request.Value.Type {
	case model.TypeGauge:
		metric.Value.Gauge = request.Value.Gauge
	case model.TypeCounter:
		metric.Value.Counter += request.Value.Counter
	}

	m.metrics[request.Name] = metric
	return nil
}

// Get возвращает метрику по имени.
func (m *MemStorage) Get(ctx context.Context, name string) (model.Metric, error) {
	metric, ok := m.metrics[name]
	if !ok {
		return model.Metric{}, fmt.Errorf("metric with name [%s] not found", name)
	}
	return metric, nil
}

// GetAll возвращает все метрики.
func (m *MemStorage) GetAll(ctx context.Context) ([]model.Metric, error) {
	metrics := make([]model.Metric, 0, len(m.metrics))
	for _, metric := range m.metrics {
		metrics = append(metrics, metric)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})
	return metrics, nil
}
