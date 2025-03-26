package apps

import (
	"github.com/NoobyTheTurtle/metrics/internal/configs"
	"github.com/NoobyTheTurtle/metrics/internal/metrics"
	"log"
	"time"
)

func StartAgent() {
	config := configs.NewAgentConfig()
	metric := metrics.NewMetrics(config.ServerAddress)

	go func() {
		for {
			time.Sleep(config.PollInterval)

			metric.UpdateMetrics()
			log.Println("Metrics updated")
		}
	}()

	go func() {
		for {
			time.Sleep(config.ReportInterval)

			metric.SendMetrics()
			log.Println("Metrics sent to server")
		}
	}()

	log.Println("Starting agent...")
	select {}
}
