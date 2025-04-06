package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemStorage_GetCounter(t *testing.T) {
	tests := []struct {
		name           string
		storage        map[string]int64
		counterName    string
		expectedValue  int64
		expectedExists bool
	}{
		{
			name: "counter exists",
			storage: map[string]int64{
				"counter1": 1,
			},
			counterName:    "counter1",
			expectedValue:  1,
			expectedExists: true,
		},
		{
			name: "counter does not exist",
			storage: map[string]int64{
				"counter1": 1,
			},
			counterName:    "counter2",
			expectedValue:  0,
			expectedExists: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				counters: tt.storage,
			}
			value, ok := ms.GetCounter(tt.counterName)

			assert.Equal(t, tt.expectedValue, value)
			assert.Equal(t, tt.expectedExists, ok)
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	tests := []struct {
		name            string
		storage         map[string]int64
		counterName     string
		updateValue     int64
		expectedStorage map[string]int64
	}{
		{
			name: "create new counter",
			storage: map[string]int64{
				"counter1": 10,
				"counter2": 20,
			},
			counterName: "counter3",
			updateValue: 5,
			expectedStorage: map[string]int64{
				"counter1": 10,
				"counter2": 20,
				"counter3": 5,
			},
		},
		{
			name: "update existing counter",
			storage: map[string]int64{
				"counter1": 10,
				"counter2": 20,
			},
			counterName: "counter1",
			updateValue: 5,
			expectedStorage: map[string]int64{
				"counter1": 15,
				"counter2": 20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				counters: tt.storage,
			}

			err := ms.UpdateCounter(tt.counterName, tt.updateValue)

			require.NoError(t, err)

			assert.Equal(t, tt.expectedStorage, ms.counters)

			value, exists := ms.GetCounter(tt.counterName)
			assert.True(t, exists)
			assert.Equal(t, tt.expectedStorage[tt.counterName], value)
		})
	}
}
