package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func (ps *PostgresStorage) Get(ctx context.Context, key string) (any, bool) {
	query := `
		SELECT value_float, value_int
		FROM metrics
		WHERE key = $1;
	`
	var valueFloat sql.NullFloat64
	var valueInt sql.NullInt64

	err := ps.db.QueryRowContext(ctx, query, key).Scan(&valueFloat, &valueInt)
	if err != nil {
		return nil, false
	}

	switch {
	case valueFloat.Valid:
		return valueFloat.Float64, true
	case valueInt.Valid:
		return valueInt.Int64, true
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

	var resultValueFloat sql.NullFloat64
	var resultValueInt sql.NullInt64

	err := ps.db.QueryRowContext(
		ctx, query, key, valueFloat, valueInt,
	).Scan(&resultValueFloat, &resultValueInt)

	if err != nil {
		return nil, fmt.Errorf("failed to set metric: %w", err)
	}

	switch {
	case resultValueFloat.Valid:
		return resultValueFloat.Float64, nil
	case resultValueInt.Valid:
		return resultValueInt.Int64, nil
	default:
		return nil, fmt.Errorf("invalid result from database")
	}
}

func (ps *PostgresStorage) GetAll(ctx context.Context) (map[string]any, error) {
	query := `
		SELECT key, value_float, value_int
		FROM metrics;
	`
	rows, err := ps.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all metrics: %w", err)
	}
	defer rows.Close()

	metrics := make(map[string]any)
	for rows.Next() {
		var key string
		var valueFloat sql.NullFloat64
		var valueInt sql.NullInt64

		if err := rows.Scan(&key, &valueFloat, &valueInt); err != nil {
			continue
		}

		switch {
		case valueFloat.Valid:
			metrics[key] = valueFloat.Float64
		case valueInt.Valid:
			metrics[key] = valueInt.Int64
		}
	}

	return metrics, nil
}
