package storage

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	value, ok := ms.counters[name]
	return value, ok
}

func (ms *MemStorage) UpdateCounter(name string, value int64) error {
	ms.counters[name] += value
	return nil
}
