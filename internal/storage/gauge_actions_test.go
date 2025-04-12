package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_GetGauge(t *testing.T) {
	tests := []struct {
		name           string
		storage        map[string]float64
		gaugeName      string
		expectedValue  float64
		expectedExists bool
	}{
		{
			name: "gauge exists",
			storage: map[string]float64{
				"gauge1": 1.1,
			},
			gaugeName:      "gauge1",
			expectedValue:  1.1,
			expectedExists: true,
		},
		{
			name: "gauge does not exist",
			storage: map[string]float64{
				"gauge1": 1.1,
			},
			gaugeName:      "gauge2",
			expectedValue:  0,
			expectedExists: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges: tt.storage,
			}
			value, ok := ms.GetGauge(tt.gaugeName)

			assert.Equal(t, tt.expectedValue, value)
			assert.Equal(t, tt.expectedExists, ok)
		})
	}
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	tests := []struct {
		name            string
		storage         map[string]float64
		gaugeName       string
		updateValue     float64
		expectedStorage map[string]float64
	}{
		{
			name: "create new gauge",
			storage: map[string]float64{
				"gauge1": 10.10,
				"gauge2": 20.20,
			},
			gaugeName:   "gauge3",
			updateValue: 5.5,
			expectedStorage: map[string]float64{
				"gauge1": 10.10,
				"gauge2": 20.20,
				"gauge3": 5.5,
			},
		},
		{
			name: "update existing gauge",
			storage: map[string]float64{
				"gauge1": 10.10,
				"gauge2": 20.20,
			},
			gaugeName:   "gauge1",
			updateValue: 5.5,
			expectedStorage: map[string]float64{
				"gauge1": 5.5,
				"gauge2": 20.20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges: tt.storage,
			}

			result, err := ms.UpdateGauge(tt.gaugeName, tt.updateValue)

			require.NoError(t, err)
			assert.Equal(t, tt.updateValue, result)
			assert.Equal(t, tt.expectedStorage, ms.gauges)

			value, exists := ms.GetGauge(tt.gaugeName)
			assert.True(t, exists)
			assert.Equal(t, tt.expectedStorage[tt.gaugeName], value)
		})
	}
}

func TestMemStorage_GetAllGauges(t *testing.T) {
	ms := &MemStorage{
		gauges: map[string]float64{
			"gauge1": 1.1,
			"gauge2": 2.2,
		},
	}

	gauges := ms.GetAllGauges()

	assert.Equal(t, ms.gauges, gauges)
}
