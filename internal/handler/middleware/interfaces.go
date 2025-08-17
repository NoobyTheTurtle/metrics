package middleware

import (
	"github.com/NoobyTheTurtle/metrics/internal/cryptoutil"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
)

type MiddlewareLogger interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
}

var (
	_ MiddlewareLogger = (*logger.ZapLogger)(nil)
	_ MiddlewareLogger = (*MockMiddlewareLogger)(nil)
)

// Decrypter определяет интерфейс для операций дешифрования
type Decrypter interface {
	Decrypt(data []byte) ([]byte, error)
}

var _ Decrypter = (*cryptoutil.PrivateKeyProvider)(nil)
