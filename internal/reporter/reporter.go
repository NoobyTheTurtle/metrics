package reporter

import (
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/metrics"
)

type Reporter struct {
	metrics        *metrics.Metrics
	logger         metrics.Logger
	reportInterval time.Duration
}

func NewReporter(metrics *metrics.Metrics, logger metrics.Logger, reportInterval uint) *Reporter {
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
