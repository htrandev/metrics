package repository

import (
	"context"
	"fmt"
	"log"
	"sort"

	models "github.com/htrandev/metrics/internal/model"
)

type MemStorage struct {
	metrics map[string]models.Metric
}

func NewMemStorageRepository() *MemStorage {
	metrics := make(map[string]models.Metric)
	return &MemStorage{
		metrics: metrics,
	}
}

func (m *MemStorage) Store(ctx context.Context, request *models.Metric) error {
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
	case models.TypeGauge:
		metric.Value.Gauge = request.Value.Gauge
	case models.TypeCounter:
		metric.Value.Counter += request.Value.Counter
	}

	m.metrics[request.Name] = metric
	return nil
}

func (m *MemStorage) Get(ctx context.Context, name string) (models.Metric, error) {
	metric, ok := m.metrics[name]
	if !ok {
		return models.Metric{}, fmt.Errorf("metric with name [%s] not found", name)
	}
	return metric, nil
}

func (m *MemStorage) GetAll(ctx context.Context) ([]models.Metric, error) {
	metrics := make([]models.Metric, 0, len(m.metrics))
	for _, metric := range m.metrics {
		metrics = append(metrics, metric)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})
	return metrics, nil
}
