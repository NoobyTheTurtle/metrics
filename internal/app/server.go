package app

import (
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/config"
	"github.com/NoobyTheTurtle/metrics/internal/handler"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
)

func StartServer() error {
	c, err := config.NewServerConfig("configs/default.yml")
	if err != nil {
		return err
	}

	isDev := c.AppEnv == "development"

	l, err := logger.NewZapLogger(c.LogLevel, isDev)
	if err != nil {
		return err
	}
	defer l.Sync()

	store := storage.NewMemStorage()

	router := handler.NewRouter(store, l)

	l.Info("Starting server on %s", c.ServerAddress)
	return http.ListenAndServe(c.ServerAddress, router.Handler())
}
