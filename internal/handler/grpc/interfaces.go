package grpc

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/database/postgres"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
)

type GaugeGetter interface {
	GetGauge(ctx context.Context, name string) (float64, bool)
}

type GaugeSetter interface {
	UpdateGauge(ctx context.Context, name string, value float64) (float64, error)
}

type CounterGetter interface {
	GetCounter(ctx context.Context, name string) (int64, bool)
}

type CounterSetter interface {
	UpdateCounter(ctx context.Context, name string, value int64) (int64, error)
}

type BatchUpdater interface {
	UpdateMetricsBatch(ctx context.Context, metrics model.Metrics) error
}

type GaugeStorage interface {
	GaugeGetter
	GaugeSetter
}

type CounterStorage interface {
	CounterGetter
	CounterSetter
}

type HandlerStorage interface {
	GaugeStorage
	CounterStorage
	BatchUpdater
}

type DBPinger interface {
	Ping(ctx context.Context) error
}

type GRPCLogger interface {
	Error(format string, args ...any)
	Info(format string, args ...any)
	Debug(format string, args ...any)
}

var (
	_ HandlerStorage = (*adapter.MetricStorage)(nil)
	_ HandlerStorage = (*MockHandlerStorage)(nil)
	_ DBPinger       = (*postgres.PostgresClient)(nil)
	_ DBPinger       = (*MockDBPinger)(nil)
	_ GRPCLogger     = (*logger.ZapLogger)(nil)
	_ GRPCLogger     = (*MockGRPCLogger)(nil)
)
