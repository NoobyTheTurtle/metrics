package postgres

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
	"github.com/NoobyTheTurtle/metrics/internal/storage/postgres/transaction"
	"github.com/jmoiron/sqlx"
)

type PostgresStorage struct {
	db *sqlx.DB
}

func NewPostgresStorage(db *sqlx.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (ps *PostgresStorage) BeginTransaction(ctx context.Context) (adapter.TransactionalStorage, error) {
	tx, err := ps.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return transaction.NewPostgresTransaction(tx), nil
}
