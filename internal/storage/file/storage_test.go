package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewFileStorage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMemStorage := NewMockMemStorage(ctrl)
	filePath := "/tmp/test.json"
	syncMode := true

	storage := NewFileStorage(mockMemStorage, filePath, syncMode)

	assert.NotNil(t, storage)
	assert.Equal(t, mockMemStorage, storage.memStorage)
	assert.Equal(t, filePath, storage.filePath)
	assert.Equal(t, syncMode, storage.syncMode)
	assert.IsType(t, &FileStorage{}, storage)
}

func TestNewFileStorage_EmptyPath_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMemStorage := NewMockMemStorage(ctrl)
	filePath := ""
	syncMode := false

	storage := NewFileStorage(mockMemStorage, filePath, syncMode)

	assert.NotNil(t, storage)
	assert.Equal(t, mockMemStorage, storage.memStorage)
	assert.Equal(t, filePath, storage.filePath)
	assert.Equal(t, syncMode, storage.syncMode)
	assert.IsType(t, &FileStorage{}, storage)
}

func TestNewFileStorage_WithDifferentSyncModes(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		syncMode bool
	}{
		{
			name:     "sync mode enabled with path",
			filePath: "/tmp/metrics.json",
			syncMode: true,
		},
		{
			name:     "sync mode disabled with path",
			filePath: "/tmp/metrics.json",
			syncMode: false,
		},
		{
			name:     "sync mode enabled with empty path",
			filePath: "",
			syncMode: true,
		},
		{
			name:     "sync mode disabled with empty path",
			filePath: "",
			syncMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMemStorage := NewMockMemStorage(ctrl)

			storage := NewFileStorage(mockMemStorage, tt.filePath, tt.syncMode)

			assert.NotNil(t, storage)
			assert.Equal(t, mockMemStorage, storage.memStorage)
			assert.Equal(t, tt.filePath, storage.filePath)
			assert.Equal(t, tt.syncMode, storage.syncMode)
		})
	}
}
