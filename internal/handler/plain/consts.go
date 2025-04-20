package plain

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"

	ContentTypeValue = "text/plain; charset=utf-8"
)
