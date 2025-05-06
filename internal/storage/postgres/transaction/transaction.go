package transaction

import (
	"github.com/jmoiron/sqlx"
)

type PostgresTransaction struct {
	tx *sqlx.Tx
}

func NewPostgresTransaction(tx *sqlx.Tx) *PostgresTransaction {
	return &PostgresTransaction{tx: tx}
}

func (pt *PostgresTransaction) Commit() error {
	return pt.tx.Commit()
}

func (pt *PostgresTransaction) Rollback() error {
	return pt.tx.Rollback()
}
