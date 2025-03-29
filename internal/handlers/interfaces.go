package handlers

import (
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

type ServerStorage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

var _ ServerStorage = (*storage.MemStorage)(nil)
var _ ServerStorage = (*storage.MockStorage)(nil)

type Logger interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
}

var _ Logger = (*logger.StdLogger)(nil)
var _ Logger = (*logger.MockLogger)(nil)
