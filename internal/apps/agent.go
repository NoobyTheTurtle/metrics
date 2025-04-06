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

	log := logger.NewStdLogger(logger.DebugLevel)
	metric := metrics.NewMetrics(config.ServerAddress, log, false)

	metricCollector := collector.NewCollector(metric, log, config.PollInterval)
	metricReporter := reporter.NewReporter(metric, log, config.ReportInterval)

	go metricCollector.Run()

	go metricReporter.Run()

	log.Info("Starting agent...")
	select {}
}
