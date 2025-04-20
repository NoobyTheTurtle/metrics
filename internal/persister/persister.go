package persister

import (
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

func (p *Persister) Run() {
	p.logger.Info("Periodic saving enabled with interval %v", p.interval)

	for {
		time.Sleep(p.interval)

		if err := p.storage.SaveToFile(); err != nil {
			p.logger.Error("Failed to save metrics: %v", err)
		} else {
			p.logger.Info("Successfully saved metrics to file")
		}
	}
}
