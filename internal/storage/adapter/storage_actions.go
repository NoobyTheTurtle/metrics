package adapter

import (
	"context"
	"fmt"
)

func (ms *MetricStorage) GetGauge(ctx context.Context, name string) (float64, bool) {
	key := addPrefix(name, GaugePrefix)
	value, exists := ms.storage.Get(ctx, key)
	if !exists {
		return 0, false
	}

	return convertToFloat64(value)
}

func (ms *MetricStorage) UpdateGauge(ctx context.Context, name string, value float64) (float64, error) {
	key := addPrefix(name, GaugePrefix)
	newValue, err := ms.storage.Set(ctx, key, value)
	if err != nil {
		return 0, fmt.Errorf("failed to update gauge metric %s: %w", name, err)
	}

	newValueFloat64, ok := convertToFloat64(newValue)
	if !ok {
		return 0, fmt.Errorf("failed to convert newValue to float64: %v", newValue)
	}

	return newValueFloat64, nil
}

func (ms *MetricStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	allMetrics, err := ms.storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all gauges: %w", err)
	}
	gauges := make(map[string]float64)

	for key, value := range allMetrics {
		if hasPrefix(key, GaugePrefix) {
			if gaugeValue, ok := convertToFloat64(value); ok {
				metricName := trimPrefix(key, GaugePrefix)
				gauges[metricName] = gaugeValue
			}
		}
	}

	return gauges, nil
}

func (ms *MetricStorage) GetCounter(ctx context.Context, name string) (int64, bool) {
	key := addPrefix(name, CounterPrefix)
	value, exists := ms.storage.Get(ctx, key)
	if !exists {
		return 0, false
	}

	return convertToInt64(value)
}

func (ms *MetricStorage) UpdateCounter(ctx context.Context, name string, value int64) (int64, error) {
	key := addPrefix(name, CounterPrefix)

	currentValue, exists := ms.GetCounter(ctx, name)
	if exists {
		value += currentValue
	}

	newValue, err := ms.storage.Set(ctx, key, value)
	if err != nil {
		return 0, fmt.Errorf("failed to update counter metric %s: %w", name, err)
	}

	newValueInt64, ok := convertToInt64(newValue)
	if !ok {
		return 0, fmt.Errorf("failed to convert newValue to int64: %v", newValue)
	}

	return newValueInt64, nil
}

func (ms *MetricStorage) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	allMetrics, err := ms.storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all counters: %w", err)
	}
	counters := make(map[string]int64)

	for key, value := range allMetrics {
		if hasPrefix(key, CounterPrefix) {
			if counterValue, ok := convertToInt64(value); ok {
				metricName := trimPrefix(key, CounterPrefix)
				counters[metricName] = counterValue
			}
		}
	}

	return counters, nil
}
