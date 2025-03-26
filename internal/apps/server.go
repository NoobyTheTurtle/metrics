package apps

import (
	"github.com/NoobyTheTurtle/metrics/internal/configs"
	"github.com/NoobyTheTurtle/metrics/internal/handlers"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
	"log"
)

func StartServer() error {
	config := configs.NewServerConfig()
	store := storage.NewMemStorage()

	log.Println("Starting server...")
	return handlers.InitHandlers(config.ServerAddress, store)
}
