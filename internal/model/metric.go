// Package model содержит основные структуры данных для сбора и обработки метрик.
package model

//go:generate easyjson -all metric.go

// MetricType определяет тип метрики.
type MetricType string

const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)

// Metric представляет одну метрику с метаданными и значением.
type Metric struct {
	ID    string     `json:"id"`              // имя метрики
	MType MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64     `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64   `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

//easyjson:json
type Metrics []Metric
