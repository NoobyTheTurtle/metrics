package postgres

import (
	"context"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PostgresClient struct {
	DB *sqlx.DB
}

func NewClient(ctx context.Context, dsn string) (*PostgresClient, error) {
	if dsn == "" {
		return &PostgresClient{DB: nil}, nil
	}

	db, err := sqlx.Open("pgx", dsn)
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
