package plain

import (
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
)

type GaugeGetter interface {
	GetGauge(name string) (float64, bool)
}

type GaugeSetter interface {
	UpdateGauge(name string, value float64) (float64, error)
}

type GaugesGetter interface {
	GetAllGauges() map[string]float64
}

type CounterGetter interface {
	GetCounter(name string) (int64, bool)
}

type CounterSetter interface {
	UpdateCounter(name string, value int64) (int64, error)
}

type CountersGetter interface {
	GetAllCounters() map[string]int64
}

type GaugeStorage interface {
	GaugeGetter
	GaugeSetter
	GaugesGetter
}

type CounterStorage interface {
	CounterGetter
	CounterSetter
	CountersGetter
}

type HandlerStorage interface {
	GaugeStorage
	CounterStorage
}

var _ HandlerStorage = (*adapter.MetricStorage)(nil)
var _ HandlerStorage = (*MockHandlerStorage)(nil)
