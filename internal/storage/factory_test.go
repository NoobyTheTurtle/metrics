package storage

import (
	"context"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/storage/adapter"
	"github.com/NoobyTheTurtle/metrics/internal/storage/file"
	"github.com/NoobyTheTurtle/metrics/internal/storage/memory"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestCreateMemoryStorage(t *testing.T) {
	storage := CreateMemoryStorage()

	assert.NotNil(t, storage)
	assert.IsType(t, &memory.MemoryStorage{}, storage)
}

func TestCreateFileStorage(t *testing.T) {
	memStorage := CreateMemoryStorage()
	filePath := "/tmp/test_metrics.json"

	storage := CreateFileStorage(memStorage, filePath, true)

	assert.NotNil(t, storage)
	assert.IsType(t, &file.FileStorage{}, storage)
}

func TestNewMetricStorage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		storageType StorageType
		filePath    string
		syncMode    bool
		restore     bool
		db          any
		expectError bool
	}{
		{
			name:        "memory storage",
			storageType: MemoryStorage,
			filePath:    "",
			syncMode:    false,
			restore:     false,
			db:          nil,
			expectError: false,
		},
		{
			name:        "file storage without restore",
			storageType: FileStorage,
			filePath:    "/tmp/test_metrics.json",
			syncMode:    true,
			restore:     false,
			db:          nil,
			expectError: false,
		},
		{
			name:        "file storage with restore",
			storageType: FileStorage,
			filePath:    "/tmp/test_metrics.json",
			syncMode:    false,
			restore:     true,
			db:          nil,
			expectError: false,
		},
		{
			name:        "unknown storage type",
			storageType: "unknown",
			filePath:    "",
			syncMode:    false,
			restore:     false,
			db:          nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var db *sqlx.DB
			if tt.db != nil {
				db = tt.db.(*sqlx.DB)
			}
			storage, err := NewMetricStorage(ctx, tt.storageType, tt.filePath, tt.syncMode, tt.restore, db)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
				assert.IsType(t, &adapter.MetricStorage{}, storage)
			}
		})
	}
}

func TestStorageTypeConstants(t *testing.T) {
	assert.Equal(t, StorageType("memory"), MemoryStorage)
	assert.Equal(t, StorageType("file"), FileStorage)
	assert.Equal(t, StorageType("postgres"), PostgresStorage)
}
