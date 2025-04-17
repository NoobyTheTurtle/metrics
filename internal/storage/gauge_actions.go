package storage

import "maps"

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, exists := ms.gauges[name]
	return value, exists
}

func (ms *MemStorage) UpdateGauge(name string, value float64) (float64, error) {
	ms.mu.Lock()
	ms.gauges[name] = value
	result := ms.gauges[name]
	ms.mu.Unlock()

	if ms.fileStoragePath != "" && ms.syncMode {
		if err := ms.SaveToFile(); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (ms *MemStorage) GetAllGauges() map[string]float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	result := make(map[string]float64, len(ms.gauges))
	maps.Copy(result, ms.gauges)
	return result
}
