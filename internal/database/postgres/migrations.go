package postgres

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (c *PostgresClient) runMigrations() error {
	if c.DB == nil {
		return errors.New("postgres.PostgresClient.runMigrations: database connection is nil")
	}

	driver, err := postgres.WithInstance(c.DB.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres.PostgresClient.runMigrations: failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/postgres",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("postgres.PostgresClient.runMigrations: failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("postgres.PostgresClient.runMigrations: failed to apply migrations: %w", err)
	}

	return nil
}
