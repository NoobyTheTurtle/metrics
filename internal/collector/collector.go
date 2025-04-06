package collector

import (
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/metrics"
)

type Collector struct {
	metrics      *metrics.Metrics
	logger       metrics.Logger
	pollInterval time.Duration
}

func NewCollector(metrics *metrics.Metrics, logger metrics.Logger, pollInterval time.Duration) *Collector {
	return &Collector{
		metrics:      metrics,
		logger:       logger,
		pollInterval: pollInterval,
	}
}

func (c *Collector) Run() {
	for {
		time.Sleep(c.pollInterval)

		c.metrics.UpdateMetrics()
		c.logger.Info("Metrics updated")
	}
}
