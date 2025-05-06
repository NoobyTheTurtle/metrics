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
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := updateMetricsBatch(ctx, tx, metrics); err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("failed to rollback transaction: %w", err)
		}
		return fmt.Errorf("failed to update metrics: %w", err)
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
				return fmt.Errorf("gauge metric %s has nil value", metric.ID)
			}

			if _, err := storage.Set(ctx, key, *metric.Value); err != nil {
				return fmt.Errorf("failed to set gauge metric: %w", err)
			}
		case model.CounterType:
			if metric.Delta == nil {
				return fmt.Errorf("counter metric %s has nil delta", metric.ID)
			}

			_, err := updateCounter(ctx, storage, metric.ID, *metric.Delta)
			if err != nil {
				return fmt.Errorf("failed to update counter metric: %w", err)
			}
		default:
			return fmt.Errorf("unknown metric type: %s", metric.MType)
		}
	}

	return nil
}
