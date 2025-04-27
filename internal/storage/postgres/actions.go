package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type Metric struct {
	Key        string          `db:"key"`
	ValueFloat sql.NullFloat64 `db:"value_float"`
	ValueInt   sql.NullInt64   `db:"value_int"`
}

func (ps *PostgresStorage) Get(ctx context.Context, key string) (any, bool) {
	query := `
		SELECT value_float, value_int
		FROM metrics
		WHERE key = $1;
	`
	var metric Metric
	err := ps.db.GetContext(ctx, &metric, query, key)
	if err != nil {
		return nil, false
	}

	switch {
	case metric.ValueFloat.Valid:
		return metric.ValueFloat.Float64, true
	case metric.ValueInt.Valid:
		return metric.ValueInt.Int64, true
	default:
		return nil, false
	}
}

func (ps *PostgresStorage) Set(ctx context.Context, key string, value any) (any, error) {
	var valueFloat sql.NullFloat64
	var valueInt sql.NullInt64

	switch v := value.(type) {
	case float64:
		valueFloat.Float64 = v
		valueFloat.Valid = true
	case int64:
		valueInt.Int64 = v
		valueInt.Valid = true
	default:
		return nil, fmt.Errorf("unsupported value type: %T", value)
	}

	query := `
		INSERT INTO metrics (key, value_float, value_int)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE
		SET value_float = $2, value_int = $3
		RETURNING value_float, value_int;
	`

	var result Metric
	err := ps.db.QueryRowxContext(
		ctx, query, key, valueFloat, valueInt,
	).StructScan(&result)

	if err != nil {
		return nil, fmt.Errorf("failed to set metric: %w", err)
	}

	switch {
	case result.ValueFloat.Valid:
		return result.ValueFloat.Float64, nil
	case result.ValueInt.Valid:
		return result.ValueInt.Int64, nil
	default:
		return nil, fmt.Errorf("invalid result from database")
	}
}

func (ps *PostgresStorage) GetAll(ctx context.Context) (map[string]any, error) {
	query := `
		SELECT key, value_float, value_int
		FROM metrics;
	`
	var metrics []Metric
	err := ps.db.SelectContext(ctx, &metrics, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all metrics: %w", err)
	}

	result := make(map[string]any, len(metrics))
	for _, m := range metrics {
		switch {
		case m.ValueFloat.Valid:
			result[m.Key] = m.ValueFloat.Float64
		case m.ValueInt.Valid:
			result[m.Key] = m.ValueInt.Int64
		}
	}

	return result, nil
}
