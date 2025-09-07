// Package grpc предоставляет gRPC сервер для API метрик.
// Реализует gRPC эндпоинты для отправки и получения метрик используя protobuf.
package grpc

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/grpc"
	"github.com/NoobyTheTurtle/metrics/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	proto.UnimplementedMetricsServiceServer
	storage HandlerStorage
	db      DBPinger
	logger  GRPCLogger
}

func NewServer(storage HandlerStorage, db DBPinger, logger GRPCLogger) *Server {
	return &Server{
		storage: storage,
		db:      db,
		logger:  logger,
	}
}

func (s *Server) UpdateMetric(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
	if req == nil || req.Metric == nil {
		return nil, status.Error(codes.InvalidArgument, "request and metric are required")
	}

	metric := req.Metric
	if metric.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "metric id is required")
	}

	switch metric.Type {
	case proto.MetricType_METRIC_TYPE_GAUGE:
		gaugeValue := metric.GetGaugeValue()
		if metric.Value == nil {
			return nil, status.Error(codes.InvalidArgument, "gauge value is required")
		}

		value, err := s.storage.UpdateGauge(ctx, metric.Id, gaugeValue)
		if err != nil {
			s.logger.Error("Failed to update gauge: %v", err)
			return nil, status.Error(codes.Internal, "failed to update gauge")
		}

		responseMetric := &proto.Metric{
			Id:   metric.Id,
			Type: proto.MetricType_METRIC_TYPE_GAUGE,
			Value: &proto.Metric_GaugeValue{
				GaugeValue: value,
			},
		}

		return &proto.UpdateMetricResponse{
			Metric: responseMetric,
		}, nil

	case proto.MetricType_METRIC_TYPE_COUNTER:
		counterDelta := metric.GetCounterDelta()
		if metric.Value == nil {
			return nil, status.Error(codes.InvalidArgument, "counter delta is required")
		}

		value, err := s.storage.UpdateCounter(ctx, metric.Id, counterDelta)
		if err != nil {
			s.logger.Error("Failed to update counter: %v", err)
			return nil, status.Error(codes.Internal, "failed to update counter")
		}

		responseMetric := &proto.Metric{
			Id:   metric.Id,
			Type: proto.MetricType_METRIC_TYPE_COUNTER,
			Value: &proto.Metric_CounterDelta{
				CounterDelta: value,
			},
		}

		return &proto.UpdateMetricResponse{
			Metric: responseMetric,
		}, nil

	case proto.MetricType_METRIC_TYPE_UNSPECIFIED:
		return nil, status.Error(codes.InvalidArgument, "metric type must be specified")

	default:
		return nil, status.Error(codes.InvalidArgument, "unknown metric type")
	}
}

func (s *Server) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if len(req.Metrics) == 0 {
		return &proto.UpdateMetricsResponse{}, nil
	}

	for _, metric := range req.Metrics {
		if metric.Id == "" {
			return nil, status.Error(codes.InvalidArgument, "metric id is required for all metrics")
		}

		switch metric.Type {
		case proto.MetricType_METRIC_TYPE_GAUGE:
			if metric.Value == nil {
				return nil, status.Error(codes.InvalidArgument, "gauge value is required")
			}
		case proto.MetricType_METRIC_TYPE_COUNTER:
			if metric.Value == nil {
				return nil, status.Error(codes.InvalidArgument, "counter delta is required")
			}
		case proto.MetricType_METRIC_TYPE_UNSPECIFIED:
			return nil, status.Error(codes.InvalidArgument, "metric type must be specified")
		default:
			return nil, status.Error(codes.InvalidArgument, "unknown metric type")
		}
	}

	internalMetrics, err := grpc.ConvertProtoMetricsToInternal(req.Metrics)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to convert metrics: "+err.Error())
	}

	err = s.storage.UpdateMetricsBatch(ctx, internalMetrics)
	if err != nil {
		s.logger.Error("Failed to update metrics batch: %v", err)
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}

	return &proto.UpdateMetricsResponse{}, nil
}

func (s *Server) GetMetric(ctx context.Context, req *proto.GetMetricRequest) (*proto.GetMetricResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "metric id is required")
	}

	switch req.Type {
	case proto.MetricType_METRIC_TYPE_GAUGE:
		value, exists := s.storage.GetGauge(ctx, req.Id)
		if !exists {
			return nil, status.Error(codes.NotFound, "gauge not found")
		}

		metric := &proto.Metric{
			Id:   req.Id,
			Type: proto.MetricType_METRIC_TYPE_GAUGE,
			Value: &proto.Metric_GaugeValue{
				GaugeValue: value,
			},
		}

		return &proto.GetMetricResponse{
			Metric: metric,
		}, nil

	case proto.MetricType_METRIC_TYPE_COUNTER:
		value, exists := s.storage.GetCounter(ctx, req.Id)
		if !exists {
			return nil, status.Error(codes.NotFound, "counter not found")
		}

		metric := &proto.Metric{
			Id:   req.Id,
			Type: proto.MetricType_METRIC_TYPE_COUNTER,
			Value: &proto.Metric_CounterDelta{
				CounterDelta: value,
			},
		}

		return &proto.GetMetricResponse{
			Metric: metric,
		}, nil

	case proto.MetricType_METRIC_TYPE_UNSPECIFIED:
		return nil, status.Error(codes.InvalidArgument, "metric type must be specified")

	default:
		return nil, status.Error(codes.InvalidArgument, "unknown metric type")
	}
}

func (s *Server) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	if s.db != nil {
		err := s.db.Ping(ctx)
		if err != nil {
			s.logger.Error("Database connection failed: %v", err)
			return nil, status.Error(codes.Unavailable, "database connection failed")
		}
	}

	return &proto.PingResponse{}, nil
}
