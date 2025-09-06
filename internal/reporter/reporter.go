package reporter

import (
	"context"
	"sync"
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

	if reportInterval == 0 {
		logger.Info("reporter.NewReporter: Invalid reportInterval value %d provided, defaulting to 1 second.", reportInterval)
		reportInterval = 1
	}

	return &Reporter{
		metrics:        metrics,
		logger:         logger,
		reportInterval: time.Duration(reportInterval) * time.Second,
		rateLimit:      rateLimit,
	}
}

func (r *Reporter) worker(workerID uint, wg *sync.WaitGroup) {
	defer wg.Done()
	r.logger.Info("Worker %d: started.", workerID)

	for range r.jobChan {
		r.metrics.SendMetrics()
		r.logger.Info("Worker %d: successfully sent metrics.", workerID)
	}

	r.logger.Info("Worker %d: stopped.", workerID)
}

func (r *Reporter) RunWithContext(ctx context.Context) {
	r.logger.Info(
		"Reporter starting with context. Report interval: %s. Worker pool size: %d.",
		r.reportInterval.String(),
		r.rateLimit,
	)

	r.jobChan = make(chan struct{}, r.rateLimit)

	var wg sync.WaitGroup
	for i := uint(1); i <= r.rateLimit; i++ {
		wg.Add(1)
		go r.worker(i, &wg)
	}

	ticker := time.NewTicker(r.reportInterval)
	defer func() {
		ticker.Stop()
		close(r.jobChan)
		r.logger.Info("Reporter waiting for workers to finish...")
		wg.Wait()
		r.logger.Info("Reporter stopped gracefully")
	}()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Reporter stopping due to context cancellation")
			return
		case <-ticker.C:
			select {
			case r.jobChan <- struct{}{}:
			case <-ctx.Done():
				r.logger.Info("Reporter stopping due to context cancellation while sending job")
				return
			}
		}
	}
}
