package collector

import (
	"context"
	"time"
)

type Collector struct {
	metrics      MetricsCollector
	logger       CollectorLogger
	pollInterval time.Duration
}

func NewCollector(metrics MetricsCollector, logger CollectorLogger, pollInterval uint) *Collector {
	if pollInterval == 0 {
		pollInterval = 1
	}

	return &Collector{
		metrics:      metrics,
		logger:       logger,
		pollInterval: time.Duration(pollInterval) * time.Second,
	}
}

func (c *Collector) RunWithContext(ctx context.Context) {
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Collector stopping due to context cancellation")
			return
		case <-ticker.C:
			c.metrics.UpdateMetrics()
			c.logger.Info("Metrics updated")
		}
	}
}
