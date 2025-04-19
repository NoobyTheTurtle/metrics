package adapter

import (
	"fmt"
)

func (ms *MetricStorage) GetGauge(name string) (float64, bool) {
	key := addPrefix(name, GaugePrefix)
	value, exists := ms.storage.Get(key)
	if !exists {
		return 0, false
	}

	return convertToFloat64(value)
}

func (ms *MetricStorage) UpdateGauge(name string, value float64) (float64, error) {
	key := addPrefix(name, GaugePrefix)
	newValue, err := ms.storage.Set(key, value)
	if err != nil {
		return 0, fmt.Errorf("failed to update gauge metric %s: %w", name, err)
	}

	newValueFloat64, ok := convertToFloat64(newValue)
	if !ok {
		return 0, fmt.Errorf("failed to convert newValue to float64: %v", newValue)
	}

	return newValueFloat64, nil
}

func (ms *MetricStorage) GetAllGauges() map[string]float64 {
	allMetrics := ms.storage.GetAll()
	gauges := make(map[string]float64)

	for key, value := range allMetrics {
		if hasPrefix(key, GaugePrefix) {
			if gaugeValue, ok := convertToFloat64(value); ok {
				metricName := trimPrefix(key, GaugePrefix)
				gauges[metricName] = gaugeValue
			}
		}
	}

	return gauges
}

func (ms *MetricStorage) GetCounter(name string) (int64, bool) {
	key := addPrefix(name, CounterPrefix)
	value, exists := ms.storage.Get(key)
	if !exists {
		return 0, false
	}

	return convertToInt64(value)
}

func (ms *MetricStorage) UpdateCounter(name string, value int64) (int64, error) {
	key := addPrefix(name, CounterPrefix)

	currentValue, exists := ms.GetCounter(name)
	if exists {
		value += currentValue
	}

	newValue, err := ms.storage.Set(key, value)
	if err != nil {
		return 0, fmt.Errorf("failed to update counter metric %s: %w", name, err)
	}

	newValueInt64, ok := convertToInt64(newValue)
	if !ok {
		return 0, fmt.Errorf("failed to convert newValue to int64: %v", newValue)
	}

	return newValueInt64, nil
}

func (ms *MetricStorage) GetAllCounters() map[string]int64 {
	allMetrics := ms.storage.GetAll()
	counters := make(map[string]int64)

	for key, value := range allMetrics {
		if hasPrefix(key, CounterPrefix) {
			if counterValue, ok := convertToInt64(value); ok {
				metricName := trimPrefix(key, CounterPrefix)
				counters[metricName] = counterValue
			}
		}
	}

	return counters
}
