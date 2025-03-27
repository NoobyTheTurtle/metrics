package storage

import "github.com/NoobyTheTurtle/metrics/internal/handlers"

var _ handlers.ServerStorage = (*MemStorage)(nil)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}
