package metric

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/cryptoutil"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/model"
)

type MetricsLogger interface {
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

var (
	_ MetricsLogger = (*logger.ZapLogger)(nil)
	_ MetricsLogger = (*MockMetricsLogger)(nil)
)

type Encrypter interface {
	Encrypt(data []byte) ([]byte, error)
}

var _ Encrypter = (*cryptoutil.PublicKeyProvider)(nil)

type MetricsTransport interface {
	SendMetrics(ctx context.Context, metrics model.Metrics) error

	Close() error
}
