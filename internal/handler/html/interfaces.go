package html

import (
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
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

var _ HandlerStorage = (*adapter.MetricStorage)(nil)
var _ HandlerStorage = (*MockHandlerStorage)(nil)
