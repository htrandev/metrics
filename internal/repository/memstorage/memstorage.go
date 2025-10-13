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

func (m *MemStorage) Get(ctx context.Context, name string) (model.Metric, error) {
	metric, ok := m.metrics[name]
	if !ok {
		return model.Metric{}, fmt.Errorf("metric with name [%s] not found", name)
	}
	return metric, nil
}

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
