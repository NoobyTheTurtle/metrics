package html

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
)

type GaugesGetter interface {
	GetAllGauges(ctx context.Context) (map[string]float64, error)
}

type CountersGetter interface {
	GetAllCounters(ctx context.Context) (map[string]int64, error)
}

type HandlerStorage interface {
	GaugesGetter
	CountersGetter
}

var _ HandlerStorage = (*adapter.MetricStorage)(nil)
var _ HandlerStorage = (*MockHandlerStorage)(nil)
