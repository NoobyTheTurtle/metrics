package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestPostgresStorage_BeginTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()

	storage := &PostgresStorage{
		db: sqlxDB,
	}

	tx, err := storage.BeginTransaction(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewPostgresStorage(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	storage := NewPostgresStorage(sqlxDB)
	assert.NotNil(t, storage)
}
