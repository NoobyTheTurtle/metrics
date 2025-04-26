package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresClient struct {
	DB *sql.DB
}

func NewClient(ctx context.Context, dsn string) (*PostgresClient, error) {
	if dsn == "" {
		return nil, errors.New("empty DSN provided")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	dbClient := &PostgresClient{DB: db}

	if err := dbClient.runMigrations(); err != nil {
		return nil, err
	}

	return dbClient, nil
}
