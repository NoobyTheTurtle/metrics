package persister

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

type PersisterLogger interface {
	Error(format string, args ...any)
	Info(format string, args ...any)
}

type MetricsStorage interface {
	SaveToFile() error
}

var _ PersisterLogger = (*logger.ZapLogger)(nil)
var _ MetricsStorage = (*storage.MemStorage)(nil)
