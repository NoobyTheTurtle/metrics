package storage

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	value, exists := ms.counters[name]
	return value, exists
}

func (ms *MemStorage) UpdateCounter(name string, value int64) error {
	ms.counters[name] += value
	return nil
}
