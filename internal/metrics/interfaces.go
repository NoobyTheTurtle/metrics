package metrics

import "github.com/NoobyTheTurtle/metrics/internal/logger"

type Logger interface {
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

var _ Logger = (*logger.StdLogger)(nil)
var _ Logger = (*logger.MockLogger)(nil)
