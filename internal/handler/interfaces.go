package handler

import (
	"github.com/NoobyTheTurtle/metrics/internal/handler/html"
	"github.com/NoobyTheTurtle/metrics/internal/handler/json"
	"github.com/NoobyTheTurtle/metrics/internal/handler/plain"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
)

type MetricStorage interface {
	html.HandlerStorage
	json.HandlerStorage
	plain.HandlerStorage
}

var _ MetricStorage = (*adapter.MetricStorage)(nil)
var _ MetricStorage = (*MockMetricStorage)(nil)

type RouterLogger interface {
	Info(format string, args ...any)
}

var _ RouterLogger = (*logger.ZapLogger)(nil)
var _ RouterLogger = (*MockRouterLogger)(nil)
