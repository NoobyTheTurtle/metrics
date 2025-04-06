package handlers

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/mocks"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

type gaugeGetter interface {
	GetGauge(name string) (float64, bool)
}

type gaugeSetter interface {
	UpdateGauge(name string, value float64) error
}

type gaugesGetter interface {
	GetAllGauges() map[string]float64
}

type counterGetter interface {
	GetCounter(name string) (int64, bool)
}

type counterSetter interface {
	UpdateCounter(name string, value int64) error
}

type countersGetter interface {
	GetAllCounters() map[string]int64
}

type gaugeStorage interface {
	gaugeGetter
	gaugeSetter
	gaugesGetter
}

type counterStorage interface {
	counterGetter
	counterSetter
	countersGetter
}

type serverStorage interface {
	gaugeStorage
	counterStorage
}

var _ serverStorage = (*storage.MemStorage)(nil)
var _ serverStorage = (*mocks.MockServerStorage)(nil)

type Logger interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
}

var _ Logger = (*logger.StdLogger)(nil)
var _ Logger = (*logger.MockLogger)(nil)
