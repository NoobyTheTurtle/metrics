package apps

import (
	"github.com/NoobyTheTurtle/metrics/internal/configs"
	"github.com/NoobyTheTurtle/metrics/internal/handlers"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

func StartServer() error {
	config := configs.NewServerConfig()
	store := storage.NewMemStorage()

	return handlers.InitHandlers(config.ServerAddress, store)
}
