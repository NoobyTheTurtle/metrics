package memory

import (
	"context"
	"testing"

	"maps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_Get(t *testing.T) {
	tests := []struct {
		name          string
		initialData   map[string]any
		key           string
		expectedValue any
		expectedFound bool
	}{
		{
			name:          "get existing value",
			initialData:   map[string]any{"test": 42},
			key:           "test",
			expectedValue: 42,
			expectedFound: true,
		},
		{
			name:          "get non-existing value",
			initialData:   map[string]any{"test": 42},
			key:           "not-exist",
			expectedValue: nil,
			expectedFound: false,
		},
		{
			name:          "empty storage",
			initialData:   map[string]any{},
			key:           "test",
			expectedValue: nil,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemoryStorage{
				data: make(map[string]any),
			}

			maps.Copy(ms.data, tt.initialData)

			ctx := context.Background()
			value, found := ms.Get(ctx, tt.key)

			assert.Equal(t, tt.expectedFound, found)
			if tt.expectedFound {
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestMemoryStorage_Set(t *testing.T) {
	tests := []struct {
		name        string
		initialData map[string]any
		key         string
		value       any
	}{
		{
			name:        "set new value",
			initialData: map[string]any{},
			key:         "test",
			value:       42,
		},
		{
			name:        "update existing value",
			initialData: map[string]any{"test": 10},
			key:         "test",
			value:       42,
		},
		{
			name:        "set with different type",
			initialData: map[string]any{"test": 10},
			key:         "test",
			value:       "string value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemoryStorage{
				data: make(map[string]any),
			}

			maps.Copy(ms.data, tt.initialData)

			ctx := context.Background()
			result, err := ms.Set(ctx, tt.key, tt.value)

			require.NoError(t, err)
			assert.Equal(t, tt.value, result)

			storedValue, exists := ms.data[tt.key]
			assert.True(t, exists)
			assert.Equal(t, tt.value, storedValue)
		})
	}
}

func TestMemoryStorage_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		initialData map[string]any
	}{
		{
			name:        "get all from populated storage",
			initialData: map[string]any{"key1": "value1", "key2": 42, "key3": true},
		},
		{
			name:        "get all from empty storage",
			initialData: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemoryStorage{
				data: make(map[string]any),
			}

			maps.Copy(ms.data, tt.initialData)

			ctx := context.Background()
			result, err := ms.GetAll(ctx)

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, len(tt.initialData), len(result))

			for key, value := range tt.initialData {
				resultValue, exists := result[key]
				assert.True(t, exists)
				assert.Equal(t, value, resultValue)
			}

			if len(result) > 0 {
				for key := range result {
					result[key] = "modified"

					ctx := context.Background()
					originalValue, _ := ms.Get(ctx, key)
					assert.NotEqual(t, "modified", originalValue)

					break
				}
			}
		})
	}
}
