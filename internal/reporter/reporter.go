package reporter

import (
	"time"
)

type Reporter struct {
	metrics        MetricsReporter
	logger         ReporterLogger
	reportInterval time.Duration
}

func NewReporter(metrics MetricsReporter, logger ReporterLogger, reportInterval uint) *Reporter {
	return &Reporter{
		metrics:        metrics,
		logger:         logger,
		reportInterval: time.Duration(reportInterval) * time.Second,
	}
}

func (r *Reporter) Run() {
	ticker := time.NewTicker(r.reportInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		r.metrics.SendMetrics()
		r.logger.Info("Metrics sent to server")
	}
}
