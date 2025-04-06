package collector

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metrics"
)

type collectorLogger interface {
	Info(format string, args ...any)
}

type metricsCollector interface {
	UpdateMetrics()
}

var _ collectorLogger = (*logger.StdLogger)(nil)
var _ metricsCollector = (*metrics.Metrics)(nil)
