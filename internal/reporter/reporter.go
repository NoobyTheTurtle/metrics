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
	for {
		time.Sleep(r.reportInterval)

		r.metrics.SendMetrics()
		r.logger.Info("Metrics sent to server")
	}
}
