// Package client предоставляет gRPC клиент для отправки метрик на сервер.
package client

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/model"
)

type MetricsClient interface {
	UpdateMetric(ctx context.Context, metric *model.Metric) (*model.Metric, error)

	UpdateMetrics(ctx context.Context, metrics model.Metrics) error

	Ping(ctx context.Context) error

	Close() error
}

type GRPCLogger interface {
	Error(format string, args ...any)
}
