package metrics

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
)

type metricsLogger interface {
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

var _ metricsLogger = (*logger.StdLogger)(nil)
var _ metricsLogger = (*MockmetricsLogger)(nil)
