package reporter

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metrics"
)

type ReporterLogger interface {
	Info(format string, args ...any)
}

type MetricsReporter interface {
	SendMetrics()
}

var _ ReporterLogger = (*logger.ZapLogger)(nil)
var _ MetricsReporter = (*metrics.Metrics)(nil)
