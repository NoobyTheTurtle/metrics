package ping

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/database/postgres"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

var _ DBPinger = (*postgres.DBClient)(nil)
var _ DBPinger = (*MockDBPinger)(nil)

type PingLogger interface {
	Error(format string, args ...any)
}

var _ PingLogger = (*logger.ZapLogger)(nil)
var _ PingLogger = (*MockPingLogger)(nil)
