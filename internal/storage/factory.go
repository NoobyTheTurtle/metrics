package storage

import (
	"fmt"

	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
	"github.com/NoobyTheTurtle/metrics/internal/storage/file"
	"github.com/NoobyTheTurtle/metrics/internal/storage/memory"
)

type StorageType string

const (
	MemoryStorage StorageType = "memory"
	FileStorage   StorageType = "file"
)

func CreateMemoryStorage() *memory.MemoryStorage {
	return memory.NewMemoryStorage()
}

func CreateFileStorage(memStorage *memory.MemoryStorage, filePath string, syncMode bool) *file.FileStorage {
	return file.NewFileStorage(memStorage, filePath, syncMode)
}

func NewMetricStorage(storageType StorageType, filePath string, syncMode bool, restore bool) (*adapter.MetricStorage, error) {
	memStorage := CreateMemoryStorage()

	var metricStorage *adapter.MetricStorage

	switch storageType {
	case FileStorage:
		fileStorage := CreateFileStorage(memStorage, filePath, syncMode)

		metricStorage = adapter.NewMetricStorage(fileStorage, fileStorage)

		if restore {
			if err := metricStorage.LoadFromFile(); err != nil {
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
