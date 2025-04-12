package html

import (
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

type GaugesGetter interface {
	GetAllGauges() map[string]float64
}

type CountersGetter interface {
	GetAllCounters() map[string]int64
}

type HandlerStorage interface {
	GaugesGetter
	CountersGetter
}

var _ HandlerStorage = (*storage.MemStorage)(nil)
var _ HandlerStorage = (*MockHandlerStorage)(nil)
