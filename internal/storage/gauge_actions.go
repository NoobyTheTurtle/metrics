package storage

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	value, exists := ms.gauges[name]
	return value, exists
}

func (ms *MemStorage) UpdateGauge(name string, value float64) error {
	ms.gauges[name] = value
	return nil
}
