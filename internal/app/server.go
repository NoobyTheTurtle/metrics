package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/config"
	"github.com/NoobyTheTurtle/metrics/internal/database/postgres"
	"github.com/NoobyTheTurtle/metrics/internal/handler"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/persister"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
	"github.com/jmoiron/sqlx"
)

func StartServer(ctx context.Context) error {
	c, err := config.NewServerConfig("configs/default.yml")
	if err != nil {
		return err
	}

	isDev := c.AppEnv == "development"

	log, err := logger.NewZapLogger(c.LogLevel, isDev)
	if err != nil {
		return err
	}
	defer log.Sync()

	dbClient, err := postgres.NewClient(ctx, c.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("app.StartServer: failed to connect to database (DSN: '%s'): %w", c.DatabaseDSN, err)
	}
	defer dbClient.Close()

	metricStorage, err := initMetricStorage(ctx, c, dbClient.DB, log)
	if err != nil {
		return fmt.Errorf("app.StartServer: failed to create metric storage: %w", err)
	}

	router := handler.NewRouter(metricStorage, log, dbClient)

	log.Info("Starting server on %s", c.ServerAddress)
	return http.ListenAndServe(c.ServerAddress, router.Handler())
}

func initMetricStorage(ctx context.Context, c *config.ServerConfig, db *sqlx.DB, log *logger.ZapLogger) (*adapter.MetricStorage, error) {
	var storageType storage.StorageType

	if c.DatabaseDSN != "" && db != nil {
		storageType = storage.PostgresStorage
	} else if c.FileStoragePath != "" {
		storageType = storage.FileStorage
	} else {
		storageType = storage.MemoryStorage
	}

	metricStorage, err := storage.NewMetricStorage(
		ctx,
		storageType,
		c.FileStoragePath,
		c.StoreInterval == 0,
		c.Restore,
		db,
	)

	if err != nil {
		return nil, err
	}

	if c.StoreInterval > 0 && storageType == storage.FileStorage {
		p := persister.NewPersister(metricStorage, log, c.StoreInterval)
		go p.Run(ctx)
	}

	return metricStorage, nil
}
