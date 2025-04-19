package memory

import "maps"

func (ms *MemoryStorage) Get(key string) (any, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.data[key]
	return value, exists
}

func (ms *MemoryStorage) Set(key string, value any) (any, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data[key] = value
	return value, nil
}

func (ms *MemoryStorage) GetAll() map[string]any {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	result := make(map[string]any, len(ms.data))
	maps.Copy(result, ms.data)
	return result
}
