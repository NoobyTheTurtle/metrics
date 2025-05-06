package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUpdateMetricsBatch(t *testing.T) {
	tests := []struct {
		name          string
		metrics       model.Metrics
		setReturnVal  any
		setError      error
		getReturnVal  any
		getFound      bool
		expectedError bool
		errorContains string
	}{
		{
			name: "successfully update gauge metrics",
			metrics: model.Metrics{
				{
					ID:    "gauge_metric",
					MType: model.GaugeType,
					Value: func() *float64 { val := 42.5; return &val }(),
				},
			},
			setReturnVal:  42.5,
			setError:      nil,
			expectedError: false,
		},
		{
			name: "successfully update counter metrics",
			metrics: model.Metrics{
				{
					ID:    "counter_metric",
					MType: model.CounterType,
					Delta: func() *int64 { val := int64(10); return &val }(),
				},
			},
			getReturnVal:  int64(5),
			getFound:      true,
			setReturnVal:  int64(15),
			setError:      nil,
			expectedError: false,
		},
		{
			name: "successfully update new counter metric",
			metrics: model.Metrics{
				{
					ID:    "new_counter",
					MType: model.CounterType,
					Delta: func() *int64 { val := int64(10); return &val }(),
				},
			},
			getReturnVal:  nil,
			getFound:      false,
			setReturnVal:  int64(10),
			setError:      nil,
			expectedError: false,
		},
		{
			name: "gauge metric with nil value",
			metrics: model.Metrics{
				{
					ID:    "nil_gauge",
					MType: model.GaugeType,
					Value: nil,
				},
			},
			expectedError: true,
			errorContains: "gauge metric nil_gauge has nil value",
		},
		{
			name: "counter metric with nil delta",
			metrics: model.Metrics{
				{
					ID:    "nil_counter",
					MType: model.CounterType,
					Delta: nil,
				},
			},
			expectedError: true,
			errorContains: "counter metric nil_counter has nil delta",
		},
		{
			name: "gauge set fails",
			metrics: model.Metrics{
				{
					ID:    "failed_gauge",
					MType: model.GaugeType,
					Value: func() *float64 { val := 42.5; return &val }(),
				},
			},
			setReturnVal:  nil,
			setError:      errors.New("storage error"),
			expectedError: true,
			errorContains: "failed to set gauge metric",
		},
		{
			name: "counter set fails",
			metrics: model.Metrics{
				{
					ID:    "failed_counter",
					MType: model.CounterType,
					Delta: func() *int64 { val := int64(10); return &val }(),
				},
			},
			getReturnVal:  int64(5),
			getFound:      true,
			setReturnVal:  nil,
			setError:      errors.New("storage error"),
			expectedError: true,
			errorContains: "failed to update counter metric",
		},
		{
			name: "unknown metric type",
			metrics: model.Metrics{
				{
					ID:    "unknown_type",
					MType: "unknown",
				},
			},
			expectedError: true,
			errorContains: "unknown metric type",
		},
		{
			name: "multiple metrics of different types",
			metrics: model.Metrics{
				{
					ID:    "gauge1",
					MType: model.GaugeType,
					Value: func() *float64 { val := 10.5; return &val }(),
				},
				{
					ID:    "counter1",
					MType: model.CounterType,
					Delta: func() *int64 { val := int64(5); return &val }(),
				},
				{
					ID:    "gauge2",
					MType: model.GaugeType,
					Value: func() *float64 { val := 42.5; return &val }(),
				},
			},
			getReturnVal:  int64(10),
			getFound:      true,
			setReturnVal:  int64(15),
			setError:      nil,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)

			ctx := context.Background()

			for _, metric := range tt.metrics {
				switch metric.MType {
				case model.GaugeType:
					if metric.Value != nil {
						mockStorage.EXPECT().
							Set(gomock.Any(), addPrefix(metric.ID, GaugePrefix), *metric.Value).
							Return(tt.setReturnVal, tt.setError).
							AnyTimes()
					}
				case model.CounterType:
					if metric.Delta != nil {
						key := addPrefix(metric.ID, CounterPrefix)
						mockStorage.EXPECT().
							Get(gomock.Any(), key).
							Return(tt.getReturnVal, tt.getFound).
							AnyTimes()

						valueToSet := *metric.Delta
						if tt.getFound {
							if current, ok := tt.getReturnVal.(int64); ok {
								valueToSet += current
							}
						}

						mockStorage.EXPECT().
							Set(gomock.Any(), key, valueToSet).
							Return(tt.setReturnVal, tt.setError).
							AnyTimes()
					}
				}
			}

			err := updateMetricsBatch(ctx, mockStorage, tt.metrics)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
