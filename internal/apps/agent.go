package apps

import (
	"github.com/NoobyTheTurtle/metrics/internal/collector"
	"github.com/NoobyTheTurtle/metrics/internal/configs"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metrics"
	"github.com/NoobyTheTurtle/metrics/internal/reporter"
)

func StartAgent() error {
	config, err := configs.NewAgentConfig()
	if err != nil {
		return err
	}

	isDev := config.AppEnv == configs.DefaultAppEnv

	l, err := logger.NewZapLogger(config.LogLevel, isDev)
	if err != nil {
		return err
	}
	defer l.Sync()

	metric := metrics.NewMetrics(config.ServerAddress, l, !isDev)

	metricCollector := collector.NewCollector(metric, l, config.PollInterval)
	metricReporter := reporter.NewReporter(metric, l, config.ReportInterval)

	go metricCollector.Run()

	go metricReporter.Run()

	l.Info("Starting agent...")
	select {}
}
