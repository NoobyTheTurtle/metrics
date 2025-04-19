package adapter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToFloat64(t *testing.T) {
	tests := []struct {
		name          string
		value         any
		expectedValue float64
		expectedOk    bool
	}{
		{
			name:          "float64 value",
			value:         float64(42.5),
			expectedValue: 42.5,
			expectedOk:    true,
		},
		{
			name:          "json.Number value",
			value:         json.Number("42.5"),
			expectedValue: 42.5,
			expectedOk:    true,
		},
		{
			name:          "int value",
			value:         42,
			expectedValue: 42.0,
			expectedOk:    true,
		},
		{
			name:          "invalid json.Number",
			value:         json.Number("not a number"),
			expectedValue: 0,
			expectedOk:    false,
		},
		{
			name:          "incompatible type",
			value:         []string{"not", "a", "number"},
			expectedValue: 0,
			expectedOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := convertToFloat64(tt.value)

			assert.Equal(t, tt.expectedOk, ok)
			if tt.expectedOk {
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestConvertToInt64(t *testing.T) {
	tests := []struct {
		name          string
		value         any
		expectedValue int64
		expectedOk    bool
	}{
		{
			name:          "int64 value",
			value:         int64(42),
			expectedValue: 42,
			expectedOk:    true,
		},
		{
			name:          "int value",
			value:         42,
			expectedValue: 42,
			expectedOk:    true,
		},
		{
			name:          "float64 value",
			value:         float64(42.5),
			expectedValue: 42,
			expectedOk:    true,
		},
		{
			name:          "json.Number value",
			value:         json.Number("42"),
			expectedValue: 42,
			expectedOk:    true,
		},
		{
			name:          "invalid json.Number",
			value:         json.Number("not a number"),
			expectedValue: 0,
			expectedOk:    false,
		},
		{
			name:          "incompatible type",
			value:         []string{"not", "a", "number"},
			expectedValue: 0,
			expectedOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := convertToInt64(tt.value)

			assert.Equal(t, tt.expectedOk, ok)
			if tt.expectedOk {
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}
