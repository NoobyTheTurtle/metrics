package reporter

import (
	"time"
)

type Reporter struct {
	metrics        MetricsReporter
	logger         ReporterLogger
	reportInterval time.Duration
	rateLimit      uint
	jobChan        chan struct{}
}

func NewReporter(
	metrics MetricsReporter,
	logger ReporterLogger,
	reportInterval uint,
	rateLimit uint,
) *Reporter {
	if rateLimit == 0 {
		logger.Info("reporter.NewReporter: Invalid rateLimit value %d provided, defaulting to 1 worker.", rateLimit)
		rateLimit = 1
	}

	return &Reporter{
		metrics:        metrics,
		logger:         logger,
		reportInterval: time.Duration(reportInterval) * time.Second,
		rateLimit:      rateLimit,
	}
}

func (r *Reporter) worker(workerID uint) {
	r.logger.Info("Worker %d: started.", workerID)

	for range r.jobChan {
		r.metrics.SendMetrics()
		r.logger.Info("Worker %d: successfully sent metrics.", workerID)
	}
}

func (r *Reporter) Run() {
	r.logger.Info(
		"Reporter starting. Report interval: %s. Worker pool size: %d.",
		r.reportInterval.String(),
		r.rateLimit,
	)

	r.jobChan = make(chan struct{}, r.rateLimit)
	defer close(r.jobChan)

	for i := uint(1); i <= r.rateLimit; i++ {
		go r.worker(i)
	}

	ticker := time.NewTicker(r.reportInterval)
	defer ticker.Stop()

	for {
		<-ticker.C

		r.jobChan <- struct{}{}
	}
}
