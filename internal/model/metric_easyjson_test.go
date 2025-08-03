package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricEasyJSON(t *testing.T) {
	floatValue := 3.14
	intValue := int64(42)

	tests := []struct {
		name      string
		metric    Metric
		expectErr bool
	}{
		{
			name: "Gauge with value",
			metric: Metric{
				ID:    "TestGauge",
				MType: GaugeType,
				Value: &floatValue,
			},
			expectErr: false,
		},
		{
			name: "Counter with delta",
			metric: Metric{
				ID:    "TestCounter",
				MType: CounterType,
				Delta: &intValue,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := tt.metric.MarshalJSON()
			assert.Equal(t, tt.expectErr, err != nil, "MarshalJSON error mismatch")
			if err == nil {
				assert.NotEmpty(t, jsonData, "MarshalJSON should not return empty data")

				var newMetric Metric
				err = newMetric.UnmarshalJSON(jsonData)
				assert.NoError(t, err, "UnmarshalJSON should not produce an error")
				assert.Equal(t, tt.metric, newMetric, "Original and unmarshaled metrics should be equal")
			}
		})
	}
}
