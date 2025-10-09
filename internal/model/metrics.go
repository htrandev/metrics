package models

import (
	"fmt"
	"strconv"
)

//easyjson:json
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type Metric struct {
	Name  string `json:"name"`
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

var metricTypeString = []string{
	"unknown",
	"gauge",
	"counter",
}

func (m MetricType) String() string {
	return metricTypeString[m]
}

func ParseMetricType(s string) MetricType {
	if v, ok := metricsTypeValues[s]; ok {
		return v
	}

	return TypeUnknown
}

type MetricValue struct {
	Type    MetricType `json:"type"`
	Gauge   float64    `json:"gauge,omitempty"`
	Counter int64      `json:"counter,omitempty"`
}

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

func (m *Metric) SetValue(s string) error {
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

func Gauge(name string, value float64) Metric {
	return Metric{
		Name:  name,
		Value: MetricValue{Type: TypeGauge, Gauge: value},
	}
}

func Counter(name string, value int64) Metric {
	return Metric{
		Name:  name,
		Value: MetricValue{Type: TypeCounter, Counter: value},
	}
}
