package apps

import (
	"github.com/NoobyTheTurtle/metrics/internal/configs"
	"github.com/NoobyTheTurtle/metrics/internal/handlers"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

func StartServer() error {
	config, err := configs.NewServerConfig()
	if err != nil {
		return err
	}

	store := storage.NewMemStorage()

	isDev := config.AppEnv == configs.DefaultAppEnv

	l, err := logger.NewZapLogger(config.LogLevel, isDev)
	if err != nil {
		return err
	}
	defer l.Sync()

	return handlers.InitHandlers(config.ServerAddress, store, l)
}
