package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
	"github.com/NoobyTheTurtle/metrics/internal/storage/file"
	"github.com/NoobyTheTurtle/metrics/internal/storage/memory"
	"github.com/NoobyTheTurtle/metrics/internal/storage/postgres"
	"github.com/jmoiron/sqlx"
)

type StorageType string

const (
	MemoryStorage   StorageType = "memory"
	FileStorage     StorageType = "file"
	PostgresStorage StorageType = "postgres"
)

func CreateMemoryStorage() *memory.MemoryStorage {
	return memory.NewMemoryStorage()
}

func CreateFileStorage(memStorage *memory.MemoryStorage, filePath string, syncMode bool) *file.FileStorage {
	return file.NewFileStorage(memStorage, filePath, syncMode)
}

func CreatePostgresStorage(db *sqlx.DB) (*postgres.PostgresStorage, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	return postgres.NewPostgresStorage(db), nil
}

func NewMetricStorage(ctx context.Context, storageType StorageType, filePath string, syncMode bool, restore bool, db *sqlx.DB) (*adapter.MetricStorage, error) {
	memStorage := CreateMemoryStorage()

	var metricStorage *adapter.MetricStorage

	switch storageType {
	case PostgresStorage:
		postgresStorage, err := CreatePostgresStorage(db)
		if err != nil {
			return nil, err
		}
		metricStorage = adapter.NewMetricStorageNoFile(postgresStorage)
	case FileStorage:
		fileStorage := CreateFileStorage(memStorage, filePath, syncMode)

		metricStorage = adapter.NewMetricStorage(fileStorage, fileStorage)

		if restore {
			if err := metricStorage.LoadFromFile(ctx); err != nil {
				return metricStorage, err
			}
		}
	case MemoryStorage:
		metricStorage = adapter.NewMetricStorageNoFile(memStorage)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}

	return metricStorage, nil
}
