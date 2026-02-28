package model

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/htrandev/metrics/internal/proto"
)

// Storager определяет интерфейс хранилища метрик.
type Storager interface {
	Get(ctx context.Context, name string) (MetricDto, error)
	GetAll(ctx context.Context) ([]MetricDto, error)

	Store(ctx context.Context, metric *MetricDto) error
	StoreMany(ctx context.Context, metrics []MetricDto) error
	StoreManyWithRetry(ctx context.Context, metrics []MetricDto) error

	Set(ctx context.Context, metric *MetricDto) error

	Ping(ctx context.Context) error
	Close() error
}

// MetricsSlice тип для массива метрик.
//
//easyjson:json
type MetricsSlice []Metrics

// Metrics содержит информацию о метрике.
//
//easyjson:json
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики.
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter.
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter.
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge.
}

// MetricDto внутренняя структура метрики с типизированным значением.
type MetricDto struct {
	Name  string `json:"name"`
	Value MetricValue
}

// MetricType тип метрики.
type MetricType uint8

const (
	TypeUnknown MetricType = iota

	TypeGauge   // Метрика типа gauge.
	TypeCounter // Метрика типа counter.
)

var metricsTypeValues = map[string]MetricType{
	"unknown": TypeUnknown,
	"gauge":   TypeGauge,
	"counter": TypeCounter,
}

var metricTypeString = []string{
	"unknown",
	"gauge",
	"counter",
}

// String возвращает строковое представление типа метрики.
func (m MetricType) String() string {
	return metricTypeString[m]
}

// ParseMetricType парсит строковый тип метрики.
func ParseMetricType(s string) MetricType {
	if v, ok := metricsTypeValues[s]; ok {
		return v
	}

	return TypeUnknown
}

// MetricValue хранит значение и тип метрики.
type MetricValue struct {
	Type    MetricType `json:"type"`
	Gauge   float64    `json:"gauge,omitempty"`
	Counter int64      `json:"counter,omitempty"`
}

// String возвращает строковое представление значения в зависимости от типа.
func (mv MetricValue) String() string {
	switch mv.Type {
	case TypeGauge:
		return strconv.FormatFloat(mv.Gauge, 'f', -1, 64)
	case TypeCounter:
		return strconv.FormatInt(mv.Counter, 10)
	default:
		return ""
	}
}

// SetValue устанавливает строковое значение метрики в соответствии с типом.
func (m *MetricDto) SetValue(s string) error {
	switch m.Value.Type {
	case TypeGauge:
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("convert value to float64: %w", err)
		}
		m.Value.Gauge = val
	case TypeCounter:
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("convert value to int64: %w", err)
		}
		m.Value.Counter = val
	}
	return nil
}

// Gauge создает новую gauge метрику с переданным значением.
func Gauge(name string, value float64) MetricDto {
	return MetricDto{
		Name:  name,
		Value: MetricValue{Type: TypeGauge, Gauge: value},
	}
}

// Gauge создает новую counter метрику с переданным значением.
func Counter(name string, value int64) MetricDto {
	return MetricDto{
		Name:  name,
		Value: MetricValue{Type: TypeCounter, Counter: value},
	}
}

func FromProto(value *pb.Metric) MetricDto {
	var m MetricDto
	switch value.GetType() {
	case pb.Metric_GAUGE:
		m = Gauge(value.GetId(), value.GetValue())
	case pb.Metric_COUNTER:
		m = Counter(value.GetId(), value.GetDelta())
	}
	return m
}

func ToProto(m MetricDto) *pb.Metric {
	var pm pb.Metric_builder

	switch m.Value.Type {
	case TypeGauge:
		pm.Id = m.Name
		pm.Type = pb.Metric_GAUGE
		pm.Value = m.Value.Gauge
	case TypeCounter:
		pm.Id = m.Name
		pm.Type = pb.Metric_COUNTER
		pm.Delta = m.Value.Counter
	}

	return pm.Build()
}
