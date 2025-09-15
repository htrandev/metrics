package models

import (
	"fmt"
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type Metric struct {
	Name  string     `json:"name"`
	Type  MetricType `json:"type"`
	Value MetricValue
}

type MetricType uint8

const (
	TypeUnknown MetricType = iota

	TypeGauge
	TypeCounter
)

var metricsTypeValues = map[string]MetricType{
	"unknown": TypeUnknown,
	"gauge":   TypeGauge,
	"counter": TypeCounter,
}

func ParseMetricType(s string) MetricType {
	if v, ok := metricsTypeValues[s]; ok {
		return v
	}

	return TypeUnknown
}

func (m *Metric) SetValue(s string) error {
	switch m.Type {
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

type MetricValue struct {
	Gauge   float64
	Counter int64
}
