package app

import (
	"github.com/NoobyTheTurtle/metrics/internal/config"
	"github.com/NoobyTheTurtle/metrics/internal/handler"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

func StartServer() error {
	c, err := config.NewServerConfig()
	if err != nil {
		return err
	}

	isDev := c.AppEnv == config.DefaultAppEnv

	l, err := logger.NewZapLogger(c.LogLevel, isDev)
	if err != nil {
		return err
	}
	defer l.Sync()

	store := storage.NewMemStorage()

	return handler.InitHandlers(c.ServerAddress, store, l)
}
