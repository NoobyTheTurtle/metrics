package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/NoobyTheTurtle/metrics/internal/retry"
)

const (
	getMetricQuery = `
		SELECT value_float, value_int
		FROM metrics
		WHERE key = $1;
	`
)

func (q *query) GetMetric(ctx context.Context, key string) (any, bool) {
	var metric Metric
	var val any
	var exists bool

	op := func() error {
		val = nil
		exists = false
		metric = Metric{}
		err := q.executor.GetContext(ctx, &metric, getMetricQuery, key)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("query.GetMetric: GetContext failed: %w", err)
		}
		switch {
		case metric.ValueFloat.Valid:
			val = metric.ValueFloat.Float64
			exists = true
		case metric.ValueInt.Valid:
			val = metric.ValueInt.Int64
			exists = true
		default:
			exists = false
		}
		return nil
	}

	retryErr := retry.WithRetries(op, retry.PgErrorChecker)
	if retryErr != nil {
		return nil, false
	}

	return val, exists
}

const (
	setMetricQuery = `
		INSERT INTO metrics (key, value_float, value_int)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE
		SET value_float = $2, value_int = $3
		RETURNING key, value_float, value_int
	`
)

func (q *query) SetMetric(ctx context.Context, key string, value any) (any, error) {
	var valueFloat sql.NullFloat64
	var valueInt sql.NullInt64
	var resultValue any

	switch v := value.(type) {
	case float64:
		valueFloat.Float64 = v
		valueFloat.Valid = true
	case int64:
		valueInt.Int64 = v
		valueInt.Valid = true
	default:
		return nil, fmt.Errorf("query.SetMetric: unsupported value type '%T'", value)
	}

	op := func() error {
		var result Metric
		resultValue = nil
		row := q.executor.QueryRowxContext(ctx, setMetricQuery, key, valueFloat, valueInt)
		if err := row.StructScan(&result); err != nil {
			return fmt.Errorf("query.SetMetric: StructScan failed: %w", err)
		}

		switch {
		case result.ValueFloat.Valid:
			resultValue = result.ValueFloat.Float64
		case result.ValueInt.Valid:
			resultValue = result.ValueInt.Int64
		default:
			return fmt.Errorf("query.SetMetric: invalid result from database")
		}
		return nil
	}

	err := retry.WithRetries(op, retry.PgErrorChecker)
	if err != nil {
		return nil, fmt.Errorf("query.SetMetric: operation failed after retries: %w", err)
	}

	return resultValue, nil
}

const (
	getAllMetricsQuery = `
		SELECT key, value_float, value_int
		FROM metrics;
	`
)

func (q *query) GetAllMetrics(ctx context.Context) (map[string]any, error) {
	var metrics []Metric
	resultData := make(map[string]any)

	op := func() error {
		metrics = []Metric{}
		resultData = make(map[string]any)
		err := q.executor.SelectContext(ctx, &metrics, getAllMetricsQuery)
		if err != nil {
			return fmt.Errorf("query.GetAllMetrics: SelectContext failed: %w", err)
		}

		for _, m := range metrics {
			switch {
			case m.ValueFloat.Valid:
				resultData[m.Key] = m.ValueFloat.Float64
			case m.ValueInt.Valid:
				resultData[m.Key] = m.ValueInt.Int64
			}
		}
		return nil
	}

	err := retry.WithRetries(op, retry.PgErrorChecker)
	if err != nil {
		return nil, fmt.Errorf("query.GetAllMetrics: operation failed after retries: %w", err)
	}

	return resultData, nil
}
