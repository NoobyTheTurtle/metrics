package postgres

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBClient struct {
	db *sql.DB
}

func NewDBClient(ctx context.Context, dsn string) (*DBClient, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DBClient{db: db}, nil
}

func (c *DBClient) Close() {
	if c.db != nil {
		c.db.Close()
	}
}

func (c *DBClient) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
