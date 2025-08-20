package app

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/collector"
	"github.com/NoobyTheTurtle/metrics/internal/config"
	"github.com/NoobyTheTurtle/metrics/internal/cryptoutil"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metric"
	"github.com/NoobyTheTurtle/metrics/internal/reporter"
)

const gracefulShutdownTimeout = 30 * time.Second

func StartAgent() error {
	c, err := config.NewAgentConfig()
	if err != nil {
		return err
	}

	isDev := c.AppEnv == "development"

	l, err := logger.NewZapLogger(c.LogLevel, isDev)
	if err != nil {
		return err
	}
	defer l.Sync()

	var encrypter metric.Encrypter
	if c.CryptoKey != "" {
		encrypter, err = cryptoutil.NewPublicKeyProvider(c.CryptoKey)
		if err != nil {
			return fmt.Errorf("app.StartAgent: failed to create encrypter: %w", err)
		}
	}

	metrics := metric.NewMetrics(c.ServerAddress, l, !isDev, c.Key, encrypter)

	metricCollector := collector.NewCollector(metrics, l, c.PollInterval)
	gopsutilCollector := collector.NewGopsutilCollector(metrics, l, c.PollInterval)
	metricReporter := reporter.NewReporter(metrics, l, c.ReportInterval, c.RateLimit)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		metricCollector.RunWithContext(ctx)
	}()

	go func() {
		defer wg.Done()
		gopsutilCollector.RunWithContext(ctx)
	}()

	go func() {
		defer wg.Done()
		metricReporter.RunWithContext(ctx)
	}()

	l.Info("Starting agent...")

	<-ctx.Done()
	l.Info("Received shutdown signal, initiating graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		l.Info("Agent stopped gracefully")
	case <-shutdownCtx.Done():
		l.Info("Graceful shutdown timed out, forcing exit")
	}

	return nil
}
