package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/config"
	"github.com/NoobyTheTurtle/metrics/internal/cryptoutil"
	"github.com/NoobyTheTurtle/metrics/internal/database/postgres"
	"github.com/NoobyTheTurtle/metrics/internal/handler"
	grpchandler "github.com/NoobyTheTurtle/metrics/internal/handler/grpc"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/persister"
	"github.com/NoobyTheTurtle/metrics/internal/storage"
	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
	"github.com/NoobyTheTurtle/metrics/proto"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
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

	httpServer := &http.Server{
		Addr:    c.ServerAddress,
		Handler: router.Handler(),
	}

	var grpcServer *grpc.Server
	if c.EnableGRPC {
		grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(grpchandler.LoggerInterceptor(log)),
		)
		grpcHandler := grpchandler.NewGRPCServer(metricStorage, dbClient, log)
		proto.RegisterMetricsServiceServer(grpcServer, grpcHandler)
	}

	serverErr := make(chan error, 2)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("Starting HTTP server on %s", c.ServerAddress)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	if c.EnableGRPC {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lis, err := net.Listen("tcp", c.GRPCServerAddress)
			if err != nil {
				serverErr <- fmt.Errorf("failed to listen on gRPC address %s: %w", c.GRPCServerAddress, err)
				return
			}
			log.Info("Starting gRPC server on %s", c.GRPCServerAddress)
			if err := grpcServer.Serve(lis); err != nil {
				serverErr <- fmt.Errorf("gRPC server error: %w", err)
			}
		}()
	}

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Info("Received shutdown signal, starting graceful shutdown...")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	log.Info("Shutting down HTTP server...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("Error during HTTP server shutdown: %v", err)
	} else {
		log.Info("HTTP server stopped")
	}

	if c.EnableGRPC && grpcServer != nil {
		log.Info("Shutting down gRPC server...")
		grpcShutdownDone := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(grpcShutdownDone)
		}()

		select {
		case <-grpcShutdownDone:
			log.Info("gRPC server stopped")
		case <-shutdownCtx.Done():
			log.Error("Timeout waiting for gRPC server shutdown, forcing stop")
			grpcServer.Stop()
		}
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
