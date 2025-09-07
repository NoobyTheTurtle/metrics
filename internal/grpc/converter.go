package grpc

import (
	"errors"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/NoobyTheTurtle/metrics/proto"
)

func ConvertInternalMetricTypeToProto(internalType model.MetricType) proto.MetricType {
	switch internalType {
	case model.GaugeType:
		return proto.MetricType_METRIC_TYPE_GAUGE
	case model.CounterType:
		return proto.MetricType_METRIC_TYPE_COUNTER
	default:
		return proto.MetricType_METRIC_TYPE_UNSPECIFIED
	}
}

func ConvertProtoMetricTypeToInternal(protoType proto.MetricType) (model.MetricType, error) {
	switch protoType {
	case proto.MetricType_METRIC_TYPE_GAUGE:
		return model.GaugeType, nil
	case proto.MetricType_METRIC_TYPE_COUNTER:
		return model.CounterType, nil
	case proto.MetricType_METRIC_TYPE_UNSPECIFIED:
		return "", errors.New("metric type must be specified")
	default:
		return "", errors.New("unknown metric type")
	}
}

func ConvertInternalMetricToProto(internalMetric *model.Metric) (*proto.Metric, error) {
	if internalMetric == nil {
		return nil, errors.New("metric cannot be nil")
	}

	if internalMetric.ID == "" {
		return nil, errors.New("metric id is required")
	}

	protoType := ConvertInternalMetricTypeToProto(internalMetric.MType)

	metric := &proto.Metric{
		Id:   internalMetric.ID,
		Type: protoType,
	}

	switch internalMetric.MType {
	case model.GaugeType:
		if internalMetric.Value == nil {
			return nil, errors.New("gauge value is required")
		}
		metric.Value = &proto.Metric_GaugeValue{
			GaugeValue: *internalMetric.Value,
		}

	case model.CounterType:
		if internalMetric.Delta == nil {
			return nil, errors.New("counter delta is required")
		}
		metric.Value = &proto.Metric_CounterDelta{
			CounterDelta: *internalMetric.Delta,
		}

	default:
		return nil, errors.New("unknown metric type")
	}

	return metric, nil
}

func ConvertProtoMetricToInternal(protoMetric *proto.Metric) (*model.Metric, error) {
	if protoMetric == nil {
		return nil, errors.New("metric cannot be nil")
	}

	if protoMetric.Id == "" {
		return nil, errors.New("metric id is required")
	}

	internalType, err := ConvertProtoMetricTypeToInternal(protoMetric.Type)
	if err != nil {
		return nil, err
	}

	metric := &model.Metric{
		ID:    protoMetric.Id,
		MType: internalType,
	}

	switch protoMetric.Type {
	case proto.MetricType_METRIC_TYPE_GAUGE:
		if protoMetric.Value == nil {
			return nil, errors.New("gauge value is required")
		}
		gaugeValue := protoMetric.GetGaugeValue()
		metric.Value = &gaugeValue

	case proto.MetricType_METRIC_TYPE_COUNTER:
		if protoMetric.Value == nil {
			return nil, errors.New("counter delta is required")
		}
		counterDelta := protoMetric.GetCounterDelta()
		metric.Delta = &counterDelta

	default:
		return nil, errors.New("unknown metric type")
	}

	return metric, nil
}

func ConvertInternalMetricsToProto(internalMetrics model.Metrics) ([]*proto.Metric, error) {
	if len(internalMetrics) == 0 {
		return []*proto.Metric{}, nil
	}

	protoMetrics := make([]*proto.Metric, len(internalMetrics))
	for i, internalMetric := range internalMetrics {
		protoMetric, err := ConvertInternalMetricToProto(&internalMetric)
		if err != nil {
			return nil, err
		}
		protoMetrics[i] = protoMetric
	}

	return protoMetrics, nil
}

func ConvertProtoMetricsToInternal(protoMetrics []*proto.Metric) (model.Metrics, error) {
	if len(protoMetrics) == 0 {
		return model.Metrics{}, nil
	}

	internalMetrics := make(model.Metrics, len(protoMetrics))
	for i, protoMetric := range protoMetrics {
		internalMetric, err := ConvertProtoMetricToInternal(protoMetric)
		if err != nil {
			return nil, err
		}
		internalMetrics[i] = *internalMetric
	}

	return internalMetrics, nil
}
