package adapter

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/storage/file"
	"github.com/NoobyTheTurtle/metrics/internal/storage/memory"
)

type Getter interface {
	Get(ctx context.Context, key string) (any, bool)
}

type Setter interface {
	Set(ctx context.Context, key string, value any) (any, error)
}

type GetAll interface {
	GetAll(ctx context.Context) (map[string]any, error)
}
type Saver interface {
	SaveToFile(ctx context.Context) error
}

type Loader interface {
	LoadFromFile(ctx context.Context) error
}

type Transaction interface {
	Commit() error
	Rollback() error
}

type TransactionalStorage interface {
	Storage
	Transaction
}

type TransactionProvider interface {
	BeginTransaction(ctx context.Context) (TransactionalStorage, error)
}

type Storage interface {
	Getter
	Setter
	GetAll
}

type FileStorage interface {
	Storage
	Saver
	Loader
}

type DatabaseStorage interface {
	Storage
	TransactionProvider
}

var _ Storage = (*memory.MemoryStorage)(nil)
var _ Storage = (*MockStorage)(nil)

var _ FileStorage = (*file.FileStorage)(nil)
var _ FileStorage = (*MockFileStorage)(nil)

// var _ DatabaseStorage = (*postgres.PostgresStorage)(nil)
var _ DatabaseStorage = (*MockDatabaseStorage)(nil)

// var _ TransactionalStorage = (*postgres.PostgresStorage)(nil)
var _ TransactionalStorage = (*MockTransactionalStorage)(nil)
