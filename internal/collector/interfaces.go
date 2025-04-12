package collector

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metrics"
)

type CollectorLogger interface {
	Info(format string, args ...any)
}

type MetricsCollector interface {
	UpdateMetrics()
}

var _ CollectorLogger = (*logger.ZapLogger)(nil)
var _ MetricsCollector = (*metrics.Metrics)(nil)
