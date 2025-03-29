package apps

import (
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/configs"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/NoobyTheTurtle/metrics/internal/metrics"
)

func StartAgent() {
	config := configs.NewAgentConfig()
	log := logger.NewStdLogger(logger.DebugLevel)
	metric := metrics.NewMetrics(config.ServerAddress, log)

	go func() {
		for {
			time.Sleep(config.PollInterval)

			metric.UpdateMetrics()
			log.Info("Metrics updated")
		}
	}()

	go func() {
		for {
			time.Sleep(config.ReportInterval)

			metric.SendMetrics()
			log.Info("Metrics sent to server")
		}
	}()

	log.Info("Starting agent...")
	select {}
}
