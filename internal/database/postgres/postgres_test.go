package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("empty DSN", func(t *testing.T) {
		ctx := context.Background()
		client, err := NewClient(ctx, "")

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Nil(t, client.DB)
	})

	t.Run("invalid DSN", func(t *testing.T) {
		ctx := context.Background()
		client, err := NewClient(ctx, "invalid://dsn")

		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("valid DSN but no connection", func(t *testing.T) {
		ctx := context.Background()
		client, err := NewClient(ctx, "postgres://user:pass@localhost:5432/nonexistent")

		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestPostgresClient_Ping(t *testing.T) {
	t.Run("nil DB", func(t *testing.T) {
		client := &PostgresClient{DB: nil}
		ctx := context.Background()

		err := client.Ping(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection is nil")
	})
}

func TestPostgresClient_Close(t *testing.T) {
	t.Run("nil DB", func(t *testing.T) {
		client := &PostgresClient{DB: nil}

		assert.NotPanics(t, func() {
			client.Close()
		})
	})
}

func TestPostgresClient_Integration(t *testing.T) {
	t.Run("client lifecycle", func(t *testing.T) {
		ctx := context.Background()

		client, err := NewClient(ctx, "")
		assert.NoError(t, err)
		assert.NotNil(t, client)

		err = client.Ping(ctx)
		assert.Error(t, err)

		assert.NotPanics(t, func() {
			client.Close()
		})
	})
}
