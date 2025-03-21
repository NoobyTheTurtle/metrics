package server

import (
	"github.com/NoobyTheTurtle/metrics/internal/handlers"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

func StartServer() error {
	store := storage.NewMemStorage()

	return handlers.InitHandlers(store)
}
