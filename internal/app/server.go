package app

import (
	"fmt"
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/config"
	"github.com/NoobyTheTurtle/metrics/internal/database/postgres"
	"github.com/NoobyTheTurtle/metrics/internal/handler"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/persister"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
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

	var dbClient *postgres.DBClient
	if c.DatabaseDSN != "" {
		dbClient, err = postgres.NewDBClient(c.DatabaseDSN)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer dbClient.Close()
	}

	metricStorage, err := getMetricStorage(c)
	if err != nil {
		return fmt.Errorf("failed to create metric storage: %w", err)
	}

	if c.StoreInterval > 0 && c.FileStoragePath != "" {
		p := persister.NewPersister(metricStorage, l, c.StoreInterval)
		go p.Run()
	}

	router := handler.NewRouter(metricStorage, l, dbClient)

	l.Info("Starting server on %s", c.ServerAddress)
	return http.ListenAndServe(c.ServerAddress, router.Handler())
}

func getMetricStorage(c *config.ServerConfig) (*adapter.MetricStorage, error) {
	var storageType storage.StorageType
	if c.FileStoragePath != "" {
		storageType = storage.FileStorage
	} else {
		storageType = storage.MemoryStorage
	}

	return storage.NewMetricStorage(
		storageType,
		c.FileStoragePath,
		c.StoreInterval == 0,
		c.Restore,
	)
}
