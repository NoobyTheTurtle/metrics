package collector

import (
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metric"
)

type CollectorLogger interface {
	Info(format string, args ...any)
}

type GopsutilMetrics interface {
	CollectGopsutilMetrics() error
	InitGopsutilMetrics(pollInterval time.Duration) error
}

type MetricsCollector interface {
	GopsutilMetrics
	UpdateMetrics()
}

var (
	_ CollectorLogger  = (*logger.ZapLogger)(nil)
	_ MetricsCollector = (*metric.Metrics)(nil)
)
