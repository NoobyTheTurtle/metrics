package handlers

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/mocks"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

type GaugeGetter interface {
	GetGauge(name string) (float64, bool)
}

type GaugeSetter interface {
	UpdateGauge(name string, value float64) error
}

type GaugesGetter interface {
	GetAllGauges() map[string]float64
}

type CounterGetter interface {
	GetCounter(name string) (int64, bool)
}

type CounterSetter interface {
	UpdateCounter(name string, value int64) error
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

type ServerStorage interface {
	GaugeStorage
	CounterStorage
}

var _ ServerStorage = (*storage.MemStorage)(nil)
var _ ServerStorage = (*mocks.MockServerStorage)(nil)

type Logger interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
}

var _ Logger = (*logger.StdLogger)(nil)
var _ Logger = (*logger.MockLogger)(nil)
