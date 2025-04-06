package reporter

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metrics"
)

type reporterLogger interface {
	Info(format string, args ...any)
}

type metricsReporter interface {
	SendMetrics()
}

var _ reporterLogger = (*logger.StdLogger)(nil)
var _ metricsReporter = (*metrics.Metrics)(nil)
