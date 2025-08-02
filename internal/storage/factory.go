// Package storage предоставляет фабрику для создания различных типов хранилищ метрик.
// Поддерживает memory, file и PostgreSQL с единым интерфейсом.
package storage

import (
	"context"
	"fmt"

	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
	"github.com/NoobyTheTurtle/metrics/internal/storage/file"
	"github.com/NoobyTheTurtle/metrics/internal/storage/memory"
	"github.com/NoobyTheTurtle/metrics/internal/storage/postgres"
	"github.com/jmoiron/sqlx"
)

// StorageType определяет тип бэкенда хранилища.
type StorageType string

const (
	MemoryStorage   StorageType = "memory"
	FileStorage     StorageType = "file"
	PostgresStorage StorageType = "postgres"
)

// CreateMemoryStorage создает хранилище в памяти.
func CreateMemoryStorage() *memory.MemoryStorage {
	return memory.NewMemoryStorage()
}

// CreateFileStorage создает файловое хранилище с бэкапом в памяти.
// syncMode определяет немедленную запись на диск.
func CreateFileStorage(memStorage *memory.MemoryStorage, filePath string, syncMode bool) *file.FileStorage {
	return file.NewFileStorage(memStorage, filePath, syncMode)
}

// CreatePostgresStorage создает PostgreSQL хранилище.
func CreatePostgresStorage(db *sqlx.DB) *postgres.PostgresStorage {
	return postgres.NewPostgresStorage(db)
}

// NewMetricStorage создает адаптер хранилища метрик по типу.
// Обрабатывает инициализацию, конфигурацию и восстановление данных для файлового хранилища.
func NewMetricStorage(ctx context.Context, storageType StorageType, filePath string, syncMode bool, restore bool, db *sqlx.DB) (*adapter.MetricStorage, error) {
	memStorage := CreateMemoryStorage()

	var metricStorage *adapter.MetricStorage

	switch storageType {
	case PostgresStorage:
		postgresStorage := CreatePostgresStorage(db)
		metricStorage = adapter.NewDatabaseStorage(postgresStorage)
	case FileStorage:
		fileStorage := CreateFileStorage(memStorage, filePath, syncMode)

		metricStorage = adapter.NewFileStorage(fileStorage)

		if restore {
			if err := metricStorage.LoadFromFile(ctx); err != nil {
				return metricStorage, err
			}
		}
	case MemoryStorage:
		metricStorage = adapter.NewStorage(memStorage)
	default:
		return nil, fmt.Errorf("storage.NewMetricStorage: unknown storage type '%s'", storageType)
	}

	return metricStorage, nil
}
