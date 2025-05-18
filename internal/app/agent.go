package app

import (
	"sync"

	"github.com/NoobyTheTurtle/metrics/internal/collector"
	"github.com/NoobyTheTurtle/metrics/internal/config"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metric"
	"github.com/NoobyTheTurtle/metrics/internal/reporter"
)

func StartAgent() error {
	c, err := config.NewAgentConfig("configs/default.yml")
	if err != nil {
		return err
	}

	isDev := c.AppEnv == "development"

	l, err := logger.NewZapLogger(c.LogLevel, isDev)
	if err != nil {
		return err
	}
	defer l.Sync()

	metrics := metric.NewMetrics(c.ServerAddress, l, !isDev, c.Key)

	metricCollector := collector.NewCollector(metrics, l, c.PollInterval)
	gopsutilCollector := collector.NewGopsutilCollector(metrics, l, c.PollInterval)
	metricReporter := reporter.NewReporter(metrics, l, c.ReportInterval)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		metricCollector.Run()
	}()

	go func() {
		defer wg.Done()
		gopsutilCollector.Run()
	}()

	go func() {
		defer wg.Done()
		metricReporter.Run()
	}()

	l.Info("Starting agent...")
	wg.Wait()
	return nil
}
