package storage

import "maps"

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	value, exists := ms.gauges[name]
	return value, exists
}

func (ms *MemStorage) UpdateGauge(name string, value float64) (float64, error) {
	ms.gauges[name] = value
	return ms.gauges[name], nil
}

func (ms *MemStorage) GetAllGauges() map[string]float64 {
	result := make(map[string]float64, len(ms.gauges))
	maps.Copy(result, ms.gauges)
	return result
}
