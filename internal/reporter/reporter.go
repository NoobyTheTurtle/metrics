package reporter

import (
	"time"
)

type Reporter struct {
	metrics        metricsReporter
	logger         reporterLogger
	reportInterval time.Duration
}

func NewReporter(metrics metricsReporter, logger reporterLogger, reportInterval uint) *Reporter {
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
