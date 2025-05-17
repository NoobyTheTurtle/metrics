package postgres

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/storage/postgres/query"
)

func (ps *PostgresStorage) Get(ctx context.Context, key string) (any, bool) {
	query := query.NewQuery(ps.db)
	return query.GetMetric(ctx, key)
}

func (ps *PostgresStorage) Set(ctx context.Context, key string, value any) (any, error) {
	query := query.NewQuery(ps.db)
	return query.SetMetric(ctx, key, value)
}

func (ps *PostgresStorage) GetAll(ctx context.Context) (map[string]any, error) {
	query := query.NewQuery(ps.db)
	return query.GetAllMetrics(ctx)
}
