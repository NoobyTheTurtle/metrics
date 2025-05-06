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
		return 0, fmt.Errorf("adapter.MetricStorage.UpdateGauge: failed to update gauge metric '%s': %w", name, err)
	}

	newValueFloat64, ok := convertToFloat64(newValue)
	if !ok {
		return 0, fmt.Errorf("adapter.MetricStorage.UpdateGauge: failed to convert newValue '%v' to float64", newValue)
	}

	return newValueFloat64, nil
}

func (ms *MetricStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	allMetrics, err := ms.storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("adapter.MetricStorage.GetAllGauges: failed to get all gauges: %w", err)
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
	if ms.dbStorage == nil {
		return updateCounter(ctx, ms.storage, name, value)
	}

	tx, err := ms.dbStorage.BeginTransaction(ctx)
	if err != nil {
		return 0, fmt.Errorf("adapter.MetricStorage.UpdateCounter: failed to begin transaction: %w", err)
	}

	value, err = updateCounter(ctx, tx, name, value)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return 0, fmt.Errorf("adapter.MetricStorage.UpdateCounter: failed to rollback transaction: %w", rollbackErr)
		}
		return 0, fmt.Errorf("adapter.MetricStorage.UpdateCounter: failed to update counter metric during transaction for '%s': %w", name, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("adapter.MetricStorage.UpdateCounter: failed to commit transaction: %w", err)
	}

	return value, nil
}

type UpdateCounterStorage interface {
	Setter
	Getter
}

func updateCounter(ctx context.Context, storage UpdateCounterStorage, name string, value int64) (int64, error) {
	key := addPrefix(name, CounterPrefix)

	currentValue, exists := storage.Get(ctx, key)
	var valueToSet = value

	if exists {
		delta, ok := convertToInt64(currentValue)
		if !ok {
			return 0, fmt.Errorf("adapter.updateCounter: failed to convert current value '%v' to int64 for key '%s'", currentValue, key)
		}

		valueToSet = value + delta
	}

	newValue, err := storage.Set(ctx, key, valueToSet)
	if err != nil {
		return 0, fmt.Errorf("adapter.updateCounter: failed to set counter metric for key '%s': %w", key, err)
	}

	newValueInt64, ok := convertToInt64(newValue)
	if !ok {
		return 0, fmt.Errorf("adapter.updateCounter: failed to convert newValue '%v' to int64 for key '%s'", newValue, key)
	}

	return newValueInt64, nil
}

func (ms *MetricStorage) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	allMetrics, err := ms.storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("adapter.MetricStorage.GetAllCounters: failed to get all counters: %w", err)
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
