package handlers

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)
