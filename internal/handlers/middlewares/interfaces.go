package middlewares

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
)

type MiddlewareLogger interface {
	Info(format string, args ...any)
}

var _ MiddlewareLogger = (*logger.ZapLogger)(nil)
var _ MiddlewareLogger = (*MockMiddlewareLogger)(nil)
