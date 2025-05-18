package collector

import (
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
	return &GopsutilCollector{
		metrics:      metrics,
		logger:       logger,
		pollInterval: time.Duration(pollInterval) * time.Second,
	}
}

func (gc *GopsutilCollector) Run() {
	ticker := time.NewTicker(gc.pollInterval)
	defer ticker.Stop()

	if err := gc.metrics.InitGomutiMetrics(gc.pollInterval); err != nil {
		gc.logger.Info(fmt.Sprintf("collector.GopsutilCollector: failed to init gopsutil metrics: %v", err))
	}

	for {
		<-ticker.C

		err := gc.metrics.CollectGopsutilMetrics()
		if err != nil {
			gc.logger.Info(fmt.Sprintf("collector.GopsutilCollector: failed to collect gopsutil metrics: %v", err))
		}
		gc.logger.Info("Gopsutil metrics updated")
	}
}
