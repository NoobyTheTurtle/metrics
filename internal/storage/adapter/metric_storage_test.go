package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewStorage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)

	ms := NewStorage(mockStorage)

	assert.NotNil(t, ms)
	assert.Equal(t, mockStorage, ms.storage)
	assert.Nil(t, ms.fileStorage)
	assert.Nil(t, ms.dbStorage)
}

func TestNewFileStorage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileStorage := NewMockFileStorage(ctrl)

	ms := NewFileStorage(mockFileStorage)

	assert.NotNil(t, ms)
	assert.Equal(t, mockFileStorage, ms.storage)
	assert.Equal(t, mockFileStorage, ms.fileStorage)
	assert.Nil(t, ms.dbStorage)
}

func TestNewDatabaseStorage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBStorage := NewMockDatabaseStorage(ctrl)

	ms := NewDatabaseStorage(mockDBStorage)

	assert.NotNil(t, ms)
	assert.Equal(t, mockDBStorage, ms.storage)
	assert.Nil(t, ms.fileStorage)
	assert.Equal(t, mockDBStorage, ms.dbStorage)
}

func TestNewDatabaseStorage_NilClient_Error(t *testing.T) {
	ms := NewDatabaseStorage(nil)

	assert.NotNil(t, ms)
	assert.Nil(t, ms.storage)
	assert.Nil(t, ms.fileStorage)
	assert.Nil(t, ms.dbStorage)
}

func TestNewStorage_NilStorage(t *testing.T) {
	ms := NewStorage(nil)

	assert.NotNil(t, ms)
	assert.Nil(t, ms.storage)
	assert.Nil(t, ms.fileStorage)
	assert.Nil(t, ms.dbStorage)
}

func TestNewFileStorage_NilFileStorage(t *testing.T) {
	ms := NewFileStorage(nil)

	assert.NotNil(t, ms)
	assert.Nil(t, ms.storage)
	assert.Nil(t, ms.fileStorage)
	assert.Nil(t, ms.dbStorage)
}

func TestMetricStorage_StorageType_Identification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name             string
		constructor      func() *MetricStorage
		expectedStorage  bool
		expectedFile     bool
		expectedDatabase bool
	}{
		{
			name: "Memory Storage",
			constructor: func() *MetricStorage {
				return NewStorage(NewMockStorage(ctrl))
			},
			expectedStorage:  true,
			expectedFile:     false,
			expectedDatabase: false,
		},
		{
			name: "File Storage",
			constructor: func() *MetricStorage {
				return NewFileStorage(NewMockFileStorage(ctrl))
			},
			expectedStorage:  true,
			expectedFile:     true,
			expectedDatabase: false,
		},
		{
			name: "Database Storage",
			constructor: func() *MetricStorage {
				return NewDatabaseStorage(NewMockDatabaseStorage(ctrl))
			},
			expectedStorage:  true,
			expectedFile:     false,
			expectedDatabase: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := tt.constructor()

			assert.Equal(t, tt.expectedStorage, ms.storage != nil, "Storage should match expected")
			assert.Equal(t, tt.expectedFile, ms.fileStorage != nil, "FileStorage should match expected")
			assert.Equal(t, tt.expectedDatabase, ms.dbStorage != nil, "DatabaseStorage should match expected")
		})
	}
}
