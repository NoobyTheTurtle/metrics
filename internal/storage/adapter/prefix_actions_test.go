package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddPrefix(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		prefix         Prefix
		expectedResult string
	}{
		{
			name:           "add gauge prefix",
			metricName:     "testMetric",
			prefix:         GaugePrefix,
			expectedResult: "gauge:testMetric",
		},
		{
			name:           "add counter prefix",
			metricName:     "testMetric",
			prefix:         CounterPrefix,
			expectedResult: "counter:testMetric",
		},
		{
			name:           "empty metric name",
			metricName:     "",
			prefix:         GaugePrefix,
			expectedResult: "gauge:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addPrefix(tt.metricName, tt.prefix)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestTrimPrefix(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		prefix         Prefix
		expectedResult string
	}{
		{
			name:           "trim gauge prefix",
			key:            "gauge:testMetric",
			prefix:         GaugePrefix,
			expectedResult: "testMetric",
		},
		{
			name:           "trim counter prefix",
			key:            "counter:testMetric",
			prefix:         CounterPrefix,
			expectedResult: "testMetric",
		},
		{
			name:           "key without prefix",
			key:            "testMetric",
			prefix:         GaugePrefix,
			expectedResult: "testMetric",
		},
		{
			name:           "empty key",
			key:            "",
			prefix:         GaugePrefix,
			expectedResult: "",
		},
		{
			name:           "only prefix",
			key:            "gauge:",
			prefix:         GaugePrefix,
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimPrefix(tt.key, tt.prefix)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		prefix         Prefix
		expectedResult bool
	}{
		{
			name:           "has gauge prefix",
			key:            "gauge:testMetric",
			prefix:         GaugePrefix,
			expectedResult: true,
		},
		{
			name:           "has counter prefix",
			key:            "counter:testMetric",
			prefix:         CounterPrefix,
			expectedResult: true,
		},
		{
			name:           "has no prefix",
			key:            "testMetric",
			prefix:         GaugePrefix,
			expectedResult: false,
		},
		{
			name:           "wrong prefix",
			key:            "gauge:testMetric",
			prefix:         CounterPrefix,
			expectedResult: false,
		},
		{
			name:           "empty key",
			key:            "",
			prefix:         GaugePrefix,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPrefix(tt.key, tt.prefix)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
