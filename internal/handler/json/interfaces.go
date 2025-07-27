package json

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
)

// GaugeGetter предоставляет чтение gauge метрик.
type GaugeGetter interface {
	GetGauge(ctx context.Context, name string) (float64, bool)
}

// GaugeSetter предоставляет запись gauge метрик.
type GaugeSetter interface {
	UpdateGauge(ctx context.Context, name string, value float64) (float64, error)
}

// CounterGetter предоставляет чтение counter метрик.
type CounterGetter interface {
	GetCounter(ctx context.Context, name string) (int64, bool)
}

// CounterSetter предоставляет запись counter метрик.
type CounterSetter interface {
	UpdateCounter(ctx context.Context, name string, value int64) (int64, error)
}

// BatchUpdater предоставляет пакетное обновление нескольких метрик.
type BatchUpdater interface {
	UpdateMetricsBatch(ctx context.Context, metrics model.Metrics) error
}

// GaugeStorage объединяет операции чтения и записи для gauge метрик.
type GaugeStorage interface {
	GaugeGetter
	GaugeSetter
}

// CounterStorage объединяет операции чтения и записи для counter метрик.
type CounterStorage interface {
	CounterGetter
	CounterSetter
}

// HandlerStorage определяет полный интерфейс хранилища для JSON обработчиков.
type HandlerStorage interface {
	GaugeStorage
	CounterStorage
	BatchUpdater
}

var (
	_ HandlerStorage = (*adapter.MetricStorage)(nil)
	_ HandlerStorage = (*MockHandlerStorage)(nil)
)
