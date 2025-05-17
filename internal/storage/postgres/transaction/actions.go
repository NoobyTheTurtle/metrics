package transaction

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/storage/postgres/query"
)

func (pt *PostgresTransaction) Get(ctx context.Context, key string) (any, bool) {
	query := query.NewQuery(pt.tx)
	return query.GetMetric(ctx, key)
}

func (pt *PostgresTransaction) Set(ctx context.Context, key string, value any) (any, error) {
	query := query.NewQuery(pt.tx)
	return query.SetMetric(ctx, key, value)
}

func (pt *PostgresTransaction) GetAll(ctx context.Context) (map[string]any, error) {
	query := query.NewQuery(pt.tx)
	return query.GetAllMetrics(ctx)
}
