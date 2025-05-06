package query

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestQuery_GetMetric(t *testing.T) {
	testutil.SkipIfNotIntegrationTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pgContainer, err := testutil.NewPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Close(ctx)

	err = pgContainer.CreateMetricsTable(ctx)
	require.NoError(t, err)

	_, err = pgContainer.DB.ExecContext(ctx,
		"INSERT INTO metrics (key, value_float, value_int) VALUES ($1, $2, $3)",
		"gauge_metric", 42.5, nil)
	require.NoError(t, err)

	_, err = pgContainer.DB.ExecContext(ctx,
		"INSERT INTO metrics (key, value_float, value_int) VALUES ($1, $2, $3)",
		"counter_metric", nil, int64(100))
	require.NoError(t, err)

	query := NewQuery(pgContainer.DB)

	t.Run("get gauge metric", func(t *testing.T) {
		value, exists := query.GetMetric(ctx, "gauge_metric")
		require.True(t, exists)
		require.Equal(t, 42.5, value)
	})

	t.Run("get counter metric", func(t *testing.T) {
		value, exists := query.GetMetric(ctx, "counter_metric")
		require.True(t, exists)
		require.Equal(t, int64(100), value)
	})

	t.Run("get non-existent metric", func(t *testing.T) {
		_, exists := query.GetMetric(ctx, "non_existent")
		require.False(t, exists)
	})
}

func TestQuery_SetMetric(t *testing.T) {
	testutil.SkipIfNotIntegrationTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pgContainer, err := testutil.NewPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Close(ctx)

	err = pgContainer.CreateMetricsTable(ctx)
	require.NoError(t, err)

	query := NewQuery(pgContainer.DB)

	t.Run("set gauge metric", func(t *testing.T) {
		result, err := query.SetMetric(ctx, "new_gauge", 123.45)
		require.NoError(t, err)
		require.Equal(t, 123.45, result)

		var metric Metric
		err = pgContainer.DB.GetContext(ctx, &metric, "SELECT key, value_float, value_int FROM metrics WHERE key = $1", "new_gauge")
		require.NoError(t, err)
		require.Equal(t, "new_gauge", metric.Key)
		require.True(t, metric.ValueFloat.Valid)
		require.Equal(t, 123.45, metric.ValueFloat.Float64)
		require.False(t, metric.ValueInt.Valid)
	})

	t.Run("set counter metric", func(t *testing.T) {
		result, err := query.SetMetric(ctx, "new_counter", int64(42))
		require.NoError(t, err)
		require.Equal(t, int64(42), result)

		var metric Metric
		err = pgContainer.DB.GetContext(ctx, &metric, "SELECT key, value_float, value_int FROM metrics WHERE key = $1", "new_counter")
		require.NoError(t, err)
		require.Equal(t, "new_counter", metric.Key)
		require.False(t, metric.ValueFloat.Valid)
		require.True(t, metric.ValueInt.Valid)
		require.Equal(t, int64(42), metric.ValueInt.Int64)
	})

	t.Run("update existing gauge metric", func(t *testing.T) {
		_, err := query.SetMetric(ctx, "update_gauge", 100.1)
		require.NoError(t, err)

		result, err := query.SetMetric(ctx, "update_gauge", 200.2)
		require.NoError(t, err)
		require.Equal(t, 200.2, result)

		var metric Metric
		err = pgContainer.DB.GetContext(ctx, &metric, "SELECT key, value_float, value_int FROM metrics WHERE key = $1", "update_gauge")
		require.NoError(t, err)
		require.Equal(t, "update_gauge", metric.Key)
		require.True(t, metric.ValueFloat.Valid)
		require.Equal(t, 200.2, metric.ValueFloat.Float64)
	})

	t.Run("set metric with unsupported type", func(t *testing.T) {
		_, err := query.SetMetric(ctx, "invalid_type", "string_value")
		require.Error(t, err)
	})
}

func TestQuery_GetAllMetrics(t *testing.T) {
	testutil.SkipIfNotIntegrationTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pgContainer, err := testutil.NewPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Close(ctx)

	err = pgContainer.CreateMetricsTable(ctx)
	require.NoError(t, err)

	data := []struct {
		key   string
		float *float64
		int   *int64
	}{
		{key: "gauge1", float: ptr(10.1), int: nil},
		{key: "gauge2", float: ptr(20.2), int: nil},
		{key: "counter1", float: nil, int: ptr(int64(100))},
		{key: "counter2", float: nil, int: ptr(int64(200))},
	}

	for _, d := range data {
		var floatVal sql.NullFloat64
		var intVal sql.NullInt64

		if d.float != nil {
			floatVal.Float64 = *d.float
			floatVal.Valid = true
		}

		if d.int != nil {
			intVal.Int64 = *d.int
			intVal.Valid = true
		}

		_, err = pgContainer.DB.ExecContext(ctx,
			"INSERT INTO metrics (key, value_float, value_int) VALUES ($1, $2, $3)",
			d.key, floatVal, intVal)
		require.NoError(t, err)
	}

	query := NewQuery(pgContainer.DB)

	t.Run("get all metrics", func(t *testing.T) {
		metrics, err := query.GetAllMetrics(ctx)
		require.NoError(t, err)
		require.Len(t, metrics, 4)

		require.Equal(t, 10.1, metrics["gauge1"])
		require.Equal(t, 20.2, metrics["gauge2"])
		require.Equal(t, int64(100), metrics["counter1"])
		require.Equal(t, int64(200), metrics["counter2"])
	})

	t.Run("get metrics from empty table", func(t *testing.T) {
		_, err := pgContainer.DB.ExecContext(ctx, "DELETE FROM metrics")
		require.NoError(t, err)

		metrics, err := query.GetAllMetrics(ctx)
		require.NoError(t, err)
		require.Empty(t, metrics)
	})
}

func ptr[T any](v T) *T {
	return &v
}
