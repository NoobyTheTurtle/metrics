package metrics

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
)

type MetricsLogger interface {
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

var _ MetricsLogger = (*logger.ZapLogger)(nil)
var _ MetricsLogger = (*MockMetricsLogger)(nil)
