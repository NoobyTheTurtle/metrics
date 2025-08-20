package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/metric"
)

type GopsutilCollector struct {
	metrics      *metric.Metrics
	logger       CollectorLogger
	pollInterval time.Duration
}

func NewGopsutilCollector(metrics *metric.Metrics, logger CollectorLogger, pollInterval uint) *GopsutilCollector {
	// Ensure poll interval is at least 1 second to avoid panic in NewTicker
	if pollInterval == 0 {
		pollInterval = 1
	}

	return &GopsutilCollector{
		metrics:      metrics,
		logger:       logger,
		pollInterval: time.Duration(pollInterval) * time.Second,
	}
}

func (gc *GopsutilCollector) RunWithContext(ctx context.Context) {
	// Check if context is already cancelled before any work
	select {
	case <-ctx.Done():
		gc.logger.Info("GopsutilCollector stopping due to context cancellation")
		return
	default:
	}

	ticker := time.NewTicker(gc.pollInterval)
	defer ticker.Stop()

	if err := gc.metrics.InitGopsutilMetrics(gc.pollInterval); err != nil {
		gc.logger.Info(fmt.Sprintf("collector.GopsutilCollector: failed to init gopsutil metrics: %v", err))
	}

	for {
		select {
		case <-ctx.Done():
			gc.logger.Info("GopsutilCollector stopping due to context cancellation")
			return
		case <-ticker.C:
			err := gc.metrics.CollectGopsutilMetrics()
			if err != nil {
				gc.logger.Info(fmt.Sprintf("collector.GopsutilCollector: failed to collect gopsutil metrics: %v", err))
			}
			gc.logger.Info("Gopsutil metrics updated")
		}
	}
}
