package storage

import "maps"

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	value, exists := ms.counters[name]
	return value, exists
}

func (ms *MemStorage) UpdateCounter(name string, value int64) error {
	ms.counters[name] += value
	return nil
}

func (ms *MemStorage) GetAllCounters() map[string]int64 {
	result := make(map[string]int64, len(ms.counters))
	maps.Copy(result, ms.counters)
	return result
}
