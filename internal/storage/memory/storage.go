package memory

import (
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]any
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]any),
	}
}
