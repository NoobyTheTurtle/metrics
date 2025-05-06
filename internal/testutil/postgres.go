package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPostgresVersion = "17-alpine"
	defaultDatabase        = "metrics_test"
	defaultUsername        = "postgres"
	defaultPassword        = "postgres"
)

type PostgresContainer struct {
	Container testcontainers.Container
	DB        *sqlx.DB
	DSN       string
}

func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	containerReq := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("postgres:%s", defaultPostgresVersion),
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       defaultDatabase,
			"POSTGRES_USER":     defaultUsername,
			"POSTGRES_PASSWORD": defaultPassword,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(time.Second * 30),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerReq,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		defaultUsername, defaultPassword, host, mappedPort.Port(), defaultDatabase)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresContainer{
		Container: container,
		DB:        db,
		DSN:       dsn,
	}, nil
}

func (pc *PostgresContainer) CreateMetricsTable(ctx context.Context) error {
	_, err := pc.DB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS metrics (
			id          SERIAL PRIMARY KEY,
			key         VARCHAR(255) NOT NULL,
			value_float DOUBLE PRECISION NULL,
			value_int   BIGINT NULL,
			UNIQUE(key)
		);
	`)
	return err
}

func (pc *PostgresContainer) Close(ctx context.Context) error {
	if pc.DB != nil {
		pc.DB.Close()
	}
	if pc.Container != nil {
		return pc.Container.Terminate(ctx)
	}
	return nil
}

func SkipIfNotIntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
}
