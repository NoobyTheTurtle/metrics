package collector

import (
	"time"
)

type Collector struct {
	metrics      MetricsCollector
	logger       CollectorLogger
	pollInterval time.Duration
}

func NewCollector(metrics MetricsCollector, logger CollectorLogger, pollInterval uint) *Collector {
	return &Collector{
		metrics:      metrics,
		logger:       logger,
		pollInterval: time.Duration(pollInterval) * time.Second,
	}
}

func (c *Collector) Run() {
	for {
		time.Sleep(c.pollInterval)

		c.metrics.UpdateMetrics()
		c.logger.Info("Metrics updated")
	}
}
