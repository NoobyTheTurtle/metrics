package handler

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/cryptoutil"
	"github.com/NoobyTheTurtle/metrics/internal/database/postgres"
	"github.com/NoobyTheTurtle/metrics/internal/handler/html"
	"github.com/NoobyTheTurtle/metrics/internal/handler/json"
	"github.com/NoobyTheTurtle/metrics/internal/handler/plain"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
)

// MetricStorage объединяет интерфейсы хранилища для всех типов обработчиков (JSON, HTML, plain text).
type MetricStorage interface {
	html.HandlerStorage
	json.HandlerStorage
	plain.HandlerStorage
}

var (
	_ MetricStorage = (*adapter.MetricStorage)(nil)
	_ MetricStorage = (*MockMetricStorage)(nil)
)

type RouterLogger interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
}

var (
	_ RouterLogger = (*logger.ZapLogger)(nil)
	_ RouterLogger = (*MockRouterLogger)(nil)
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

var (
	_ DBPinger = (*postgres.PostgresClient)(nil)
	_ DBPinger = (*MockDBPinger)(nil)
)

// Decrypter определяет интерфейс для операций дешифрования
type Decrypter interface {
	Decrypt(data []byte) ([]byte, error)
}

var _ Decrypter = (*cryptoutil.PrivateKeyProvider)(nil)
