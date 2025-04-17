package storage

import (
	"sync"
)

type MemStorage struct {
	mu              sync.RWMutex
	gauges          map[string]float64
	counters        map[string]int64
	fileStoragePath string
	syncMode        bool
}

func NewMemStorage(fileStoragePath string, syncMode bool) *MemStorage {
	return &MemStorage{
		gauges:          make(map[string]float64),
		counters:        make(map[string]int64),
		fileStoragePath: fileStoragePath,
		syncMode:        syncMode,
	}
}
