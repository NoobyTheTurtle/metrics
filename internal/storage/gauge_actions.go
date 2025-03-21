package storage

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	value, ok := ms.gauges[name]
	return value, ok
}

func (ms *MemStorage) UpdateGauge(name string, value float64) error {
	ms.gauges[name] = value
	return nil
}
