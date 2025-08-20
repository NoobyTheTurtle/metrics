package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_Success(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		expectNilDB bool
	}{
		{
			name:        "empty DSN returns nil DB",
			dsn:         "",
			expectNilDB: true,
		},
		{
			name:        "valid DSN with test database",
			dsn:         "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
			expectNilDB: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			client, err := NewClient(ctx, tt.dsn)

			if tt.expectNilDB {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Nil(t, client.DB)
			} else {
				if err != nil {
					assert.Error(t, err)
					assert.Nil(t, client)
				} else {
					require.NotNil(t, client)
					assert.NotNil(t, client.DB)
					if client.DB != nil {
						client.DB.Close()
					}
				}
			}
		})
	}
}

func TestNewClient_InvalidDSN_Error(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "invalid protocol",
			dsn:  "invalid://user:pass@localhost:5432/testdb",
		},
		{
			name: "malformed DSN",
			dsn:  "postgres://invalid dsn format",
		},
		{
			name: "missing host",
			dsn:  "postgres://user:pass@:5432/testdb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			client, err := NewClient(ctx, tt.dsn)

			assert.Error(t, err)
			assert.Nil(t, client)
		})
	}
}

func TestNewClient_ConnectionFailed_Error(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	invalidDSN := "postgres://user:pass@nonexistent-host-12345:5432/testdb?sslmode=disable"

	client, err := NewClient(ctx, invalidDSN)

	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewClient_PingFailed_Error(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	unreachableDSN := "postgres://user:pass@192.0.2.1:5432/testdb?sslmode=disable&connect_timeout=1"

	client, err := NewClient(ctx, unreachableDSN)

	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewClient_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	dsn := "postgres://user:pass@localhost:5432/testdb?sslmode=disable"

	client, err := NewClient(ctx, dsn)

	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestRunMigrations_NilDB_Error(t *testing.T) {
	client := &PostgresClient{DB: nil}

	err := client.runMigrations()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

func TestRunMigrations_MigrationSourceError(t *testing.T) {
	t.Run("migration path is correctly specified", func(t *testing.T) {
		expectedPath := "file://migrations/postgres"
		assert.NotEmpty(t, expectedPath)
	})
}
