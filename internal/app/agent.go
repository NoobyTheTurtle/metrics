package app

import (
	"github.com/NoobyTheTurtle/metrics/internal/collector"
	"github.com/NoobyTheTurtle/metrics/internal/config"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metric"
	"github.com/NoobyTheTurtle/metrics/internal/reporter"
)

func StartAgent() error {
	c, err := config.NewAgentConfig()
	if err != nil {
		return err
	}

	isDev := c.AppEnv == config.DefaultAppEnv

	l, err := logger.NewZapLogger(c.LogLevel, isDev)
	if err != nil {
		return err
	}
	defer l.Sync()

	metrics := metric.NewMetrics(c.ServerAddress, l, !isDev)

	metricCollector := collector.NewCollector(metrics, l, c.PollInterval)
	metricReporter := reporter.NewReporter(metrics, l, c.ReportInterval)

	go metricCollector.Run()

	go metricReporter.Run()

	l.Info("Starting agent...")
	select {}
}
