package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFileStorage_Get(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		mockValue     any
		mockFound     bool
		expectedValue any
		expectedFound bool
	}{
		{
			name:          "get existing value",
			key:           "test",
			mockValue:     42,
			mockFound:     true,
			expectedValue: 42,
			expectedFound: true,
		},
		{
			name:          "get non-existing value",
			key:           "not-exist",
			mockValue:     nil,
			mockFound:     false,
			expectedValue: nil,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMemStorage := NewMockMemStorage(ctrl)
			mockMemStorage.EXPECT().
				Get(tt.key).
				Return(tt.mockValue, tt.mockFound)

			fs := &FileStorage{
				memStorage: mockMemStorage,
			}

			value, found := fs.Get(tt.key)

			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

func TestFileStorage_Set(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		key      string
		value    any
		syncMode bool
		filePath string
	}{
		{
			name:     "set value without sync",
			key:      "test",
			value:    42,
			syncMode: false,
			filePath: "",
		},
		{
			name:     "set value with sync",
			key:      "test",
			value:    42,
			syncMode: true,
			filePath: filepath.Join(tempDir, "test_sync.json"),
		},
		{
			name:     "set value with empty path",
			key:      "test",
			value:    42,
			syncMode: true,
			filePath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMemStorage := NewMockMemStorage(ctrl)
			mockMemStorage.EXPECT().
				Set(tt.key, tt.value).
				Return(tt.value, nil)

			if tt.syncMode && tt.filePath != "" {
				mockMemStorage.EXPECT().
					GetAll().
					Return(map[string]any{tt.key: tt.value})
			}

			fs := &FileStorage{
				memStorage: mockMemStorage,
				syncMode:   tt.syncMode,
				filePath:   tt.filePath,
			}

			result, err := fs.Set(tt.key, tt.value)

			assert.NoError(t, err)
			assert.Equal(t, tt.value, result)

			if tt.syncMode && tt.filePath != "" {
				fileExists := false
				_, err := os.Stat(tt.filePath)
				fileExists = !os.IsNotExist(err)
				assert.True(t, fileExists)

				fileData, readErr := os.ReadFile(tt.filePath)
				require.NoError(t, readErr)

				var savedData map[string]any
				unmarshalErr := json.Unmarshal(fileData, &savedData)
				require.NoError(t, unmarshalErr)

				assert.Contains(t, savedData, tt.key)

				if intValue, ok := tt.value.(int); ok {
					assert.Equal(t, float64(intValue), savedData[tt.key])
				} else {
					assert.Equal(t, tt.value, savedData[tt.key])
				}
			}
		})
	}
}

func TestFileStorage_Set_SaveError(t *testing.T) {
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")
	require.NoError(t, os.Mkdir(readOnlyDir, 0500))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMemStorage := NewMockMemStorage(ctrl)
	mockMemStorage.EXPECT().
		Set("test", 42).
		Return(42, nil)
	mockMemStorage.EXPECT().
		GetAll().
		Return(map[string]any{"test": 42})

	fs := &FileStorage{
		memStorage: mockMemStorage,
		syncMode:   true,
		filePath:   filepath.Join(readOnlyDir, "cannot_write.json"),
	}

	_, err := fs.Set("test", 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save to file")
}

func TestFileStorage_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		mockData    map[string]any
		expectedLen int
	}{
		{
			name:        "get all from populated storage",
			mockData:    map[string]any{"key1": "value1", "key2": 42},
			expectedLen: 2,
		},
		{
			name:        "get all from empty storage",
			mockData:    map[string]any{},
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMemStorage := NewMockMemStorage(ctrl)
			mockMemStorage.EXPECT().
				GetAll().
				Return(tt.mockData)

			fs := &FileStorage{
				memStorage: mockMemStorage,
			}

			result := fs.GetAll()

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedLen, len(result))

			for key, value := range tt.mockData {
				resultValue, exists := result[key]
				assert.True(t, exists)
				assert.Equal(t, value, resultValue)
			}
		})
	}
}

func TestFileStorage_SaveToFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		data     map[string]any
		filePath string
	}{
		{
			name:     "save to valid file path",
			data:     map[string]any{"key1": "value1", "key2": 42},
			filePath: filepath.Join(tempDir, "test.json"),
		},
		{
			name:     "save with empty file path",
			data:     map[string]any{"key1": "value1", "key2": 42},
			filePath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMemStorage := NewMockMemStorage(ctrl)
			mockMemStorage.EXPECT().
				GetAll().
				Return(tt.data).
				AnyTimes()

			fs := &FileStorage{
				memStorage: mockMemStorage,
				filePath:   tt.filePath,
			}

			err := fs.SaveToFile()

			assert.NoError(t, err)

			if tt.filePath != "" {
				fileData, readErr := os.ReadFile(tt.filePath)
				require.NoError(t, readErr)

				var savedData map[string]any
				unmarshalErr := json.Unmarshal(fileData, &savedData)
				require.NoError(t, unmarshalErr)

				expectedJSON, err := json.Marshal(tt.data)
				require.NoError(t, err)
				actualJSON, err := json.Marshal(savedData)
				require.NoError(t, err)
				assert.JSONEq(t, string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func TestFileStorage_LoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.json")

	testData := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	jsonData, err := json.Marshal(testData)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tempFile, jsonData, 0644))

	tests := []struct {
		name          string
		filePath      string
		fileExists    bool
		expectedCalls int
	}{
		{
			name:          "load from existing file",
			filePath:      tempFile,
			fileExists:    true,
			expectedCalls: len(testData),
		},
		{
			name:          "load from non-existent file",
			filePath:      filepath.Join(tempDir, "nonexistent.json"),
			fileExists:    false,
			expectedCalls: 0,
		},
		{
			name:          "load with empty file path",
			filePath:      "",
			fileExists:    false,
			expectedCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMemStorage := NewMockMemStorage(ctrl)

			if tt.fileExists {
				mockMemStorage.EXPECT().
					Set("key1", "value1").
					Return("value1", nil)
				mockMemStorage.EXPECT().
					Set("key2", float64(42)).
					Return(float64(42), nil)
			}

			fs := &FileStorage{
				memStorage: mockMemStorage,
				filePath:   tt.filePath,
			}

			err := fs.LoadFromFile()

			assert.NoError(t, err)
		})
	}
}

func TestFileStorage_LoadFromFile_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "invalid.json")

	require.NoError(t, os.WriteFile(tempFile, []byte("{invalid json}"), 0644))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMemStorage := NewMockMemStorage(ctrl)

	fs := &FileStorage{
		memStorage: mockMemStorage,
		filePath:   tempFile,
	}

	err := fs.LoadFromFile()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal data")
}

func TestFileStorage_LoadFromFile_SetError(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.json")

	testData := map[string]any{
		"key1": "value1",
	}
	jsonData, err := json.Marshal(testData)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tempFile, jsonData, 0644))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMemStorage := NewMockMemStorage(ctrl)
	mockMemStorage.EXPECT().
		Set("key1", "value1").
		Return(nil, assert.AnError)

	fs := &FileStorage{
		memStorage: mockMemStorage,
		filePath:   tempFile,
	}

	err = fs.LoadFromFile()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set value for key")
}
