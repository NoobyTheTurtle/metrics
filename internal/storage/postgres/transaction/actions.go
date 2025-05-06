package transaction

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/storage/postgres/query"
)

func (ps *PostgresTransaction) Get(ctx context.Context, key string) (any, bool) {
	query := query.NewQuery(ps.tx)
	return query.GetMetric(ctx, key)
}

func (ps *PostgresTransaction) Set(ctx context.Context, key string, value any) (any, error) {
	query := query.NewQuery(ps.tx)
	return query.SetMetric(ctx, key, value)
}

func (ps *PostgresTransaction) GetAll(ctx context.Context) (map[string]any, error) {
	query := query.NewQuery(ps.tx)
	return query.GetAllMetrics(ctx)
}
