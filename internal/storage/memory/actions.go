package memory

import (
	"context"
	"maps"
)

func (ms *MemoryStorage) Get(ctx context.Context, key string) (any, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.data[key]
	return value, exists
}

func (ms *MemoryStorage) Set(ctx context.Context, key string, value any) (any, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data[key] = value
	return ms.data[key], nil
}

func (ms *MemoryStorage) GetAll(ctx context.Context) (map[string]any, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	result := make(map[string]any, len(ms.data))
	maps.Copy(result, ms.data)
	return result, nil
}
