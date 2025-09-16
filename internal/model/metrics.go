package models

import (
	"fmt"
	"strconv"
)

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

type MetricValue struct {
	Gauge   float64
	Counter int64
}

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

func Gauge(name string, value float64) Metric {
	return Metric{
		Name:  name,
		Type:  TypeGauge,
		Value: MetricValue{Gauge: value},
	}
}

func Counter(name string, value int64) Metric {
	return Metric{
		Name:  name,
		Type:  TypeCounter,
		Value: MetricValue{Counter: value},
	}
}
