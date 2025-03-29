package apps

import (
	"github.com/NoobyTheTurtle/metrics/internal/configs"
	"github.com/NoobyTheTurtle/metrics/internal/handlers"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

func StartServer() error {
	config := configs.NewServerConfig()
	store := storage.NewMemStorage()
	log := logger.NewStdLogger(logger.DebugLevel)

	log.Info("Starting server...")
	return handlers.InitHandlers(config.ServerAddress, store, log)
}
