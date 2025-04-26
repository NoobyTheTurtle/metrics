package json

import (
	"context"

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
}

var _ HandlerStorage = (*adapter.MetricStorage)(nil)
var _ HandlerStorage = (*MockHandlerStorage)(nil)
