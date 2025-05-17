package adapter

import (
	"context"
	"fmt"

	"github.com/NoobyTheTurtle/metrics/internal/model"
)

func (ms *MetricStorage) UpdateMetricsBatch(ctx context.Context, metrics model.Metrics) error {
	if ms.dbStorage == nil {
		return updateMetricsBatch(ctx, ms.storage, metrics)
	}

	tx, err := ms.dbStorage.BeginTransaction(ctx)
	if err != nil {
		return fmt.Errorf("adapter.MetricStorage.UpdateMetricsBatch: failed to begin transaction: %w", err)
	}

	if err := updateMetricsBatch(ctx, tx, metrics); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("adapter.MetricStorage.UpdateMetricsBatch: failed to rollback transaction: %w", rollbackErr)
		}
		return fmt.Errorf("adapter.MetricStorage.UpdateMetricsBatch: failed to update metrics batch: %w", err)
	}

	return tx.Commit()
}

type BatchStorage interface {
	Setter
	Getter
}

func updateMetricsBatch(ctx context.Context, storage BatchStorage, metrics model.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case model.GaugeType:
			key := addPrefix(metric.ID, GaugePrefix)

			if metric.Value == nil {
				return fmt.Errorf("adapter.updateMetricsBatch: gauge metric '%s' has nil value", metric.ID)
			}

			if _, err := storage.Set(ctx, key, *metric.Value); err != nil {
				return fmt.Errorf("adapter.updateMetricsBatch: failed to set gauge metric '%s': %w", metric.ID, err)
			}
		case model.CounterType:
			if metric.Delta == nil {
				return fmt.Errorf("adapter.updateMetricsBatch: counter metric '%s' has nil delta", metric.ID)
			}

			_, err := updateCounter(ctx, storage, metric.ID, *metric.Delta)
			if err != nil {
				return fmt.Errorf("adapter.updateMetricsBatch: failed to update counter metric '%s': %w", metric.ID, err)
			}
		default:
			return fmt.Errorf("adapter.updateMetricsBatch: unknown metric type '%s' for metric ID '%s'", metric.MType, metric.ID)
		}
	}

	return nil
}
