package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestInitGopsutilMetrics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockMetricsLogger(ctrl)
	metrics := NewMetrics("localhost:8080", mockLogger, false, "", nil)

	pollInterval := 100 * time.Millisecond

	err := metrics.InitGopsutilMetrics(pollInterval)

	assert.NoError(t, err)
}

func TestInitGopsutilMetrics_ZeroPollInterval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockMetricsLogger(ctrl)
	metrics := NewMetrics("localhost:8080", mockLogger, false, "", nil)

	err := metrics.InitGopsutilMetrics(0)

	assert.NoError(t, err)
}

func TestCollectGopsutilMetrics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockMetricsLogger(ctrl)
	metrics := NewMetrics("localhost:8080", mockLogger, false, "", nil)

	err := metrics.CollectGopsutilMetrics()

	require.NoError(t, err)

	totalMemory, exists := metrics.Gauges[GaugeMetric("TotalMemory")]
	assert.True(t, exists, "TotalMemory should be collected")
	assert.Greater(t, totalMemory, 0.0, "TotalMemory should be positive")

	freeMemory, exists := metrics.Gauges[GaugeMetric("FreeMemory")]
	assert.True(t, exists, "FreeMemory should be collected")
	assert.GreaterOrEqual(t, freeMemory, 0.0, "FreeMemory should be non-negative")

	assert.LessOrEqual(t, freeMemory, totalMemory, "FreeMemory should be less than or equal to TotalMemory")

	foundCPUMetrics := false
	for gaugeKey := range metrics.Gauges {
		if gaugeKey == GaugeMetric("CPUutilization1") {
			foundCPUMetrics = true
			cpuValue := metrics.Gauges[gaugeKey]
			assert.GreaterOrEqual(t, cpuValue, 0.0, "CPU utilization should be non-negative")
			assert.LessOrEqual(t, cpuValue, 100.0, "CPU utilization should not exceed 100%")
			break
		}
	}
	assert.True(t, foundCPUMetrics, "At least CPUutilization1 should be collected")
}

func TestCollectGopsutilMetrics_MultipleCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockMetricsLogger(ctrl)
	metrics := NewMetrics("localhost:8080", mockLogger, false, "", nil)

	err1 := metrics.CollectGopsutilMetrics()
	require.NoError(t, err1)

	firstTotalMemory := metrics.Gauges[GaugeMetric("TotalMemory")]

	err2 := metrics.CollectGopsutilMetrics()
	require.NoError(t, err2)

	secondTotalMemory := metrics.Gauges[GaugeMetric("TotalMemory")]

	assert.Equal(t, firstTotalMemory, secondTotalMemory, "TotalMemory should remain consistent across calls")
}

func TestCollectGopsutilMetrics_MetricTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockMetricsLogger(ctrl)
	metrics := NewMetrics("localhost:8080", mockLogger, false, "", nil)

	err := metrics.CollectGopsutilMetrics()
	require.NoError(t, err)

	tests := []struct {
		name       string
		metricName GaugeMetric
		required   bool
	}{
		{"TotalMemory", GaugeMetric("TotalMemory"), true},
		{"FreeMemory", GaugeMetric("FreeMemory"), true},
		{"CPUutilization1", GaugeMetric("CPUutilization1"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists := metrics.Gauges[tt.metricName]
			if tt.required {
				assert.True(t, exists, "Required metric %s should exist", tt.metricName)
				assert.GreaterOrEqual(t, value, 0.0, "Metric %s should be non-negative", tt.metricName)
			}
		})
	}
}
