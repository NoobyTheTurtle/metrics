package persister

import (
	"context"
	"time"
)

type Persister struct {
	storage  MetricsStorage
	logger   PersisterLogger
	interval time.Duration
}

func NewPersister(storage MetricsStorage, logger PersisterLogger, storeInterval uint) *Persister {
	return &Persister{
		storage:  storage,
		logger:   logger,
		interval: time.Duration(storeInterval) * time.Second,
	}
}

func (p *Persister) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.logger.Info("Periodic saving enabled with interval %v", p.interval)

	for {
		<-ticker.C
		if err := p.storage.SaveToFile(ctx); err != nil {
			p.logger.Error("Failed to save metrics: %v", err)
		} else {
			p.logger.Info("Successfully saved metrics to file")
		}
	}
}
