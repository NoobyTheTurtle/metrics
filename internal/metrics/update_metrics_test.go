package metrics

import (
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_UpdateMetrics(t *testing.T) {
	metrics := NewMetrics("localhost:8080", logger.NewMockLogger())

	_, exists := metrics.Gauges[HeapObjects]
	assert.False(t, exists, "HeapObjects should not exist before update")

	_, exists = metrics.Gauges[RandomValue]
	assert.False(t, exists, "RandomValue should not exist before update")

	_, exists = metrics.Counters[PollCount]
	assert.False(t, exists, "PollCount should not exist before update")

	metrics.UpdateMetrics()

	_, exists = metrics.Gauges[HeapObjects]
	assert.True(t, exists, "HeapObjects should exist after update")

	_, exists = metrics.Gauges[RandomValue]
	assert.True(t, exists, "RandomValue should exist after update")

	pollCount, exists := metrics.Counters[PollCount]
	assert.True(t, exists, "PollCount should exist after update")
	assert.Equal(t, int64(1), pollCount, "PollCount should be incremented to 1")
}

func TestMetrics_updateGaugeMemStats(t *testing.T) {
	metrics := NewMetrics("localhost:8080", logger.NewMockLogger())

	metrics.updateGaugeMemStats()

	requiredGauges := []GaugeMetric{
		Alloc, BuckHashSys, Frees, GCCPUFraction, GCSys, HeapAlloc,
		HeapIdle, HeapInuse, HeapObjects, HeapReleased, HeapSys,
		LastGC, Lookups, MCacheInuse, MCacheSys, MSpanInuse,
		MSpanSys, Mallocs, NextGC, NumForcedGC, NumGC, OtherSys,
		PauseTotalNs, StackInuse, StackSys, Sys, TotalAlloc,
	}

	for _, metricName := range requiredGauges {
		_, exists := metrics.Gauges[metricName]
		assert.True(t, exists, "%s should exist after updateGaugeMemStats", metricName)
	}

	_, exists := metrics.Gauges[RandomValue]
	assert.False(t, exists, "RandomValue should not be set by updateGaugeMemStats")
}

func TestMetrics_updateGaugeRandomValue(t *testing.T) {
	metrics := NewMetrics("localhost:8080", logger.NewMockLogger())

	metrics.updateGaugeRandomValue()

	randomValue, exists := metrics.Gauges[RandomValue]
	assert.True(t, exists, "RandomValue should exist after updateGaugeRandomValue")
	assert.GreaterOrEqual(t, randomValue, 0.0, "RandomValue should be >= 0.0")
	assert.Less(t, randomValue, 1.0, "RandomValue should be < 1.0")

	firstValue := randomValue
	metrics.updateGaugeRandomValue()
	secondValue, exists := metrics.Gauges[RandomValue]
	assert.True(t, exists, "RandomValue should still exist after second updateGaugeRandomValue")

	assert.NotEqual(t, firstValue, secondValue, "RandomValue should change between calls")
}

func TestMetrics_updateCounters(t *testing.T) {
	tests := []struct {
		name              string
		initialPollCount  int64
		expectedPollCount int64
	}{
		{
			name:              "increment from zero",
			initialPollCount:  0,
			expectedPollCount: 1,
		},
		{
			name:              "increment from positive value",
			initialPollCount:  42,
			expectedPollCount: 43,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics("localhost:8080", logger.NewMockLogger())

			if tt.initialPollCount > 0 {
				metrics.Counters[PollCount] = tt.initialPollCount
			}

			metrics.updateCounters()

			pollCount, exists := metrics.Counters[PollCount]
			assert.True(t, exists, "PollCount should exist after updateCounters")
			assert.Equal(t, tt.expectedPollCount, pollCount, "PollCount should be incremented correctly")
		})
	}
}
