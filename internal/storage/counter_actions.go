package storage

import "maps"

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.counters[name]
	return value, exists
}

func (ms *MemStorage) UpdateCounter(name string, value int64) (int64, error) {
	ms.mu.Lock()
	ms.counters[name] += value
	result := ms.counters[name]
	ms.mu.Unlock()

	if ms.fileStoragePath != "" && ms.syncMode {
		if err := ms.SaveToFile(); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (ms *MemStorage) GetAllCounters() map[string]int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	result := make(map[string]int64, len(ms.counters))
	maps.Copy(result, ms.counters)
	return result
}
