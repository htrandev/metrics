package repository

import (
	"log"

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

func (m *MemStorage) Store(request *models.Metric) error {
	if request == nil {
		log.Println("repository: request is nil")
		return nil
	}
	metric, ok := m.metrics[request.Name]
	if !ok {
		m.metrics[request.Name] = *request
		return nil
	}

	switch request.Type {
	case models.TypeGauge:
		metric.Value.Gauge = request.Value.Gauge
	case models.TypeCounter:
		metric.Value.Counter += request.Value.Counter
	}

	m.metrics[request.Name] = metric
	return nil
}
