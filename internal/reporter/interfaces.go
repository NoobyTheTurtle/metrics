package reporter

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metric"
)

type ReporterLogger interface {
	Info(format string, args ...any)
}

type MetricsReporter interface {
	SendMetrics()
}

var _ ReporterLogger = (*logger.ZapLogger)(nil)
var _ MetricsReporter = (*metric.Metrics)(nil)
