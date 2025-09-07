package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/NoobyTheTurtle/metrics/proto"
)

func TestConvertInternalMetricTypeToProto(t *testing.T) {
	testCases := []struct {
		name          string
		internalType  model.MetricType
		expectedProto proto.MetricType
	}{
		{
			name:          "gauge type",
			internalType:  model.GaugeType,
			expectedProto: proto.MetricType_METRIC_TYPE_GAUGE,
		},
		{
			name:          "counter type",
			internalType:  model.CounterType,
			expectedProto: proto.MetricType_METRIC_TYPE_COUNTER,
		},
		{
			name:          "unknown type",
			internalType:  "unknown",
			expectedProto: proto.MetricType_METRIC_TYPE_UNSPECIFIED,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertInternalMetricTypeToProto(tc.internalType)
			assert.Equal(t, tc.expectedProto, result)
		})
	}
}

func TestConvertProtoMetricTypeToInternal(t *testing.T) {
	testCases := []struct {
		name             string
		protoType        proto.MetricType
		expectedInternal model.MetricType
		expectError      bool
	}{
		{
			name:             "gauge type",
			protoType:        proto.MetricType_METRIC_TYPE_GAUGE,
			expectedInternal: model.GaugeType,
			expectError:      false,
		},
		{
			name:             "counter type",
			protoType:        proto.MetricType_METRIC_TYPE_COUNTER,
			expectedInternal: model.CounterType,
			expectError:      false,
		},
		{
			name:        "unspecified type",
			protoType:   proto.MetricType_METRIC_TYPE_UNSPECIFIED,
			expectError: true,
		},
		{
			name:        "unknown type",
			protoType:   proto.MetricType(999),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ConvertProtoMetricTypeToInternal(tc.protoType)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedInternal, result)
			}
		})
	}
}

func TestConvertInternalMetricToProto(t *testing.T) {
	testCases := []struct {
		name        string
		metric      *model.Metric
		expectError bool
		expected    *proto.Metric
	}{
		{
			name:        "nil metric",
			metric:      nil,
			expectError: true,
		},
		{
			name: "empty id",
			metric: &model.Metric{
				ID:    "",
				MType: model.GaugeType,
				Value: func() *float64 { v := 1.0; return &v }(),
			},
			expectError: true,
		},
		{
			name: "gauge with value",
			metric: &model.Metric{
				ID:    "test_gauge",
				MType: model.GaugeType,
				Value: func() *float64 { v := 42.5; return &v }(),
			},
			expectError: false,
			expected: &proto.Metric{
				Id:   "test_gauge",
				Type: proto.MetricType_METRIC_TYPE_GAUGE,
				Value: &proto.Metric_GaugeValue{
					GaugeValue: 42.5,
				},
			},
		},
		{
			name: "gauge with nil value",
			metric: &model.Metric{
				ID:    "test_gauge",
				MType: model.GaugeType,
				Value: nil,
			},
			expectError: true,
		},
		{
			name: "counter with delta",
			metric: &model.Metric{
				ID:    "test_counter",
				MType: model.CounterType,
				Delta: func() *int64 { v := int64(10); return &v }(),
			},
			expectError: false,
			expected: &proto.Metric{
				Id:   "test_counter",
				Type: proto.MetricType_METRIC_TYPE_COUNTER,
				Value: &proto.Metric_CounterDelta{
					CounterDelta: 10,
				},
			},
		},
		{
			name: "counter with nil delta",
			metric: &model.Metric{
				ID:    "test_counter",
				MType: model.CounterType,
				Delta: nil,
			},
			expectError: true,
		},
		{
			name: "unknown type",
			metric: &model.Metric{
				ID:    "test_unknown",
				MType: "unknown",
				Value: func() *float64 { v := 1.0; return &v }(),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ConvertInternalMetricToProto(tc.metric)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestConvertProtoMetricToInternal(t *testing.T) {
	testCases := []struct {
		name        string
		metric      *proto.Metric
		expectError bool
		expected    *model.Metric
	}{
		{
			name:        "nil metric",
			metric:      nil,
			expectError: true,
		},
		{
			name: "empty id",
			metric: &proto.Metric{
				Id:   "",
				Type: proto.MetricType_METRIC_TYPE_GAUGE,
				Value: &proto.Metric_GaugeValue{
					GaugeValue: 42.5,
				},
			},
			expectError: true,
		},
		{
			name: "gauge with value",
			metric: &proto.Metric{
				Id:   "test_gauge",
				Type: proto.MetricType_METRIC_TYPE_GAUGE,
				Value: &proto.Metric_GaugeValue{
					GaugeValue: 42.5,
				},
			},
			expectError: false,
			expected: &model.Metric{
				ID:    "test_gauge",
				MType: model.GaugeType,
				Value: func() *float64 { v := 42.5; return &v }(),
			},
		},
		{
			name: "gauge with nil value",
			metric: &proto.Metric{
				Id:    "test_gauge",
				Type:  proto.MetricType_METRIC_TYPE_GAUGE,
				Value: nil,
			},
			expectError: true,
		},
		{
			name: "counter with delta",
			metric: &proto.Metric{
				Id:   "test_counter",
				Type: proto.MetricType_METRIC_TYPE_COUNTER,
				Value: &proto.Metric_CounterDelta{
					CounterDelta: 10,
				},
			},
			expectError: false,
			expected: &model.Metric{
				ID:    "test_counter",
				MType: model.CounterType,
				Delta: func() *int64 { v := int64(10); return &v }(),
			},
		},
		{
			name: "counter with nil delta",
			metric: &proto.Metric{
				Id:    "test_counter",
				Type:  proto.MetricType_METRIC_TYPE_COUNTER,
				Value: nil,
			},
			expectError: true,
		},
		{
			name: "unspecified type",
			metric: &proto.Metric{
				Id:   "test_unspecified",
				Type: proto.MetricType_METRIC_TYPE_UNSPECIFIED,
				Value: &proto.Metric_GaugeValue{
					GaugeValue: 1.0,
				},
			},
			expectError: true,
		},
		{
			name: "unknown type",
			metric: &proto.Metric{
				Id:   "test_unknown",
				Type: proto.MetricType(999),
				Value: &proto.Metric_GaugeValue{
					GaugeValue: 1.0,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ConvertProtoMetricToInternal(tc.metric)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestConvertInternalMetricsToProto(t *testing.T) {
	testCases := []struct {
		name        string
		metrics     model.Metrics
		expectError bool
		expected    []*proto.Metric
	}{
		{
			name:        "empty metrics",
			metrics:     model.Metrics{},
			expectError: false,
			expected:    []*proto.Metric{},
		},
		{
			name: "single gauge metric",
			metrics: model.Metrics{
				{
					ID:    "gauge1",
					MType: model.GaugeType,
					Value: func() *float64 { v := 1.5; return &v }(),
				},
			},
			expectError: false,
			expected: []*proto.Metric{
				{
					Id:   "gauge1",
					Type: proto.MetricType_METRIC_TYPE_GAUGE,
					Value: &proto.Metric_GaugeValue{
						GaugeValue: 1.5,
					},
				},
			},
		},
		{
			name: "multiple metrics",
			metrics: model.Metrics{
				{
					ID:    "gauge1",
					MType: model.GaugeType,
					Value: func() *float64 { v := 1.5; return &v }(),
				},
				{
					ID:    "counter1",
					MType: model.CounterType,
					Delta: func() *int64 { v := int64(5); return &v }(),
				},
			},
			expectError: false,
			expected: []*proto.Metric{
				{
					Id:   "gauge1",
					Type: proto.MetricType_METRIC_TYPE_GAUGE,
					Value: &proto.Metric_GaugeValue{
						GaugeValue: 1.5,
					},
				},
				{
					Id:   "counter1",
					Type: proto.MetricType_METRIC_TYPE_COUNTER,
					Value: &proto.Metric_CounterDelta{
						CounterDelta: 5,
					},
				},
			},
		},
		{
			name: "error in one metric",
			metrics: model.Metrics{
				{
					ID:    "gauge1",
					MType: model.GaugeType,
					Value: func() *float64 { v := 1.5; return &v }(),
				},
				{
					ID:    "",
					MType: model.CounterType,
					Delta: func() *int64 { v := int64(5); return &v }(),
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ConvertInternalMetricsToProto(tc.metrics)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestConvertProtoMetricsToInternal(t *testing.T) {
	testCases := []struct {
		name        string
		metrics     []*proto.Metric
		expectError bool
		expected    model.Metrics
	}{
		{
			name:        "empty metrics",
			metrics:     []*proto.Metric{},
			expectError: false,
			expected:    model.Metrics{},
		},
		{
			name: "single gauge metric",
			metrics: []*proto.Metric{
				{
					Id:   "gauge1",
					Type: proto.MetricType_METRIC_TYPE_GAUGE,
					Value: &proto.Metric_GaugeValue{
						GaugeValue: 1.5,
					},
				},
			},
			expectError: false,
			expected: model.Metrics{
				{
					ID:    "gauge1",
					MType: model.GaugeType,
					Value: func() *float64 { v := 1.5; return &v }(),
				},
			},
		},
		{
			name: "multiple metrics",
			metrics: []*proto.Metric{
				{
					Id:   "gauge1",
					Type: proto.MetricType_METRIC_TYPE_GAUGE,
					Value: &proto.Metric_GaugeValue{
						GaugeValue: 1.5,
					},
				},
				{
					Id:   "counter1",
					Type: proto.MetricType_METRIC_TYPE_COUNTER,
					Value: &proto.Metric_CounterDelta{
						CounterDelta: 5,
					},
				},
			},
			expectError: false,
			expected: model.Metrics{
				{
					ID:    "gauge1",
					MType: model.GaugeType,
					Value: func() *float64 { v := 1.5; return &v }(),
				},
				{
					ID:    "counter1",
					MType: model.CounterType,
					Delta: func() *int64 { v := int64(5); return &v }(),
				},
			},
		},
		{
			name: "error in one metric",
			metrics: []*proto.Metric{
				{
					Id:   "gauge1",
					Type: proto.MetricType_METRIC_TYPE_GAUGE,
					Value: &proto.Metric_GaugeValue{
						GaugeValue: 1.5,
					},
				},
				{
					Id:   "",
					Type: proto.MetricType_METRIC_TYPE_COUNTER,
					Value: &proto.Metric_CounterDelta{
						CounterDelta: 5,
					},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ConvertProtoMetricsToInternal(tc.metrics)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
