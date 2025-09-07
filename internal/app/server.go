package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/config"
	"github.com/NoobyTheTurtle/metrics/internal/cryptoutil"
	"github.com/NoobyTheTurtle/metrics/internal/database/postgres"
	"github.com/NoobyTheTurtle/metrics/internal/handler"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/persister"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
	"github.com/jmoiron/sqlx"
)

func StartServer(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	c, err := config.NewServerConfig()
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

	metricStorage, persisterDone, err := initMetricStorage(ctx, c, dbClient.DB, log)
	if err != nil {
		return fmt.Errorf("app.StartServer: failed to create metric storage: %w", err)
	}

	var decrypter handler.Decrypter
	if c.CryptoKey != "" {
		decrypter, err = cryptoutil.NewPrivateKeyProvider(c.CryptoKey)
		if err != nil {
			return fmt.Errorf("app.StartServer: failed to create decrypter: %w", err)
		}
	}

	router := handler.NewRouter(metricStorage, log, dbClient, c.Key, decrypter, c.TrustedSubnet)

	server := &http.Server{
		Addr:    c.ServerAddress,
		Handler: router.Handler(),
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("Starting server on %s", c.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Info("Received shutdown signal, starting graceful shutdown...")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	log.Info("Shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Error during HTTP server shutdown: %v", err)
	} else {
		log.Info("HTTP server stopped")
	}

	if persisterDone != nil {
		log.Info("Waiting for persister to finish...")
		select {
		case <-persisterDone:
			log.Info("Persister finished")
		case <-shutdownCtx.Done():
			log.Error("Timeout waiting for persister to finish")
		}
	}

	log.Info("Graceful shutdown completed")
	return nil
}

func initMetricStorage(ctx context.Context, c *config.ServerConfig, db *sqlx.DB, log *logger.ZapLogger) (*adapter.MetricStorage, chan struct{}, error) {
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
		return nil, nil, err
	}

	var persisterDone chan struct{}
	if c.StoreInterval > 0 && storageType == storage.FileStorage {
		persisterDone = make(chan struct{})
		p := persister.NewPersister(metricStorage, log, c.StoreInterval)
		go func() {
			defer close(persisterDone)
			p.Run(ctx)
		}()
	}

	return metricStorage, persisterDone, nil
}
