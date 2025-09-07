package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/NoobyTheTurtle/metrics/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockHandlerStorage(ctrl)
	mockDB := NewMockDBPinger(ctrl)
	mockLogger := NewMockGRPCLogger(ctrl)

	server := NewServer(mockStorage, mockDB, mockLogger)

	assert.NotNil(t, server)
	assert.Equal(t, mockStorage, server.storage)
	assert.Equal(t, mockDB, server.db)
	assert.Equal(t, mockLogger, server.logger)
}

func TestServer_UpdateMetric(t *testing.T) {
	tests := []struct {
		name        string
		request     *proto.UpdateMetricRequest
		setupMocks  func(*MockHandlerStorage, *MockGRPCLogger)
		expectedErr codes.Code
		validate    func(*testing.T, *proto.UpdateMetricResponse)
	}{
		{
			name: "successful gauge update",
			request: &proto.UpdateMetricRequest{
				Metric: &proto.Metric{
					Id:   "HeapObjects",
					Type: proto.MetricType_METRIC_TYPE_GAUGE,
					Value: &proto.Metric_GaugeValue{
						GaugeValue: 7770.0,
					},
				},
			},
			setupMocks: func(storage *MockHandlerStorage, logger *MockGRPCLogger) {
				storage.EXPECT().UpdateGauge(gomock.Any(), "HeapObjects", 7770.0).Return(7770.0, nil)
			},
			expectedErr: codes.OK,
			validate: func(t *testing.T, resp *proto.UpdateMetricResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Metric)
				assert.Equal(t, "HeapObjects", resp.Metric.Id)
				assert.Equal(t, proto.MetricType_METRIC_TYPE_GAUGE, resp.Metric.Type)
				assert.Equal(t, 7770.0, resp.Metric.GetGaugeValue())
			},
		},
		{
			name: "successful counter update",
			request: &proto.UpdateMetricRequest{
				Metric: &proto.Metric{
					Id:   "PollCount",
					Type: proto.MetricType_METRIC_TYPE_COUNTER,
					Value: &proto.Metric_CounterDelta{
						CounterDelta: 30,
					},
				},
			},
			setupMocks: func(storage *MockHandlerStorage, logger *MockGRPCLogger) {
				storage.EXPECT().UpdateCounter(gomock.Any(), "PollCount", int64(30)).Return(int64(30), nil)
			},
			expectedErr: codes.OK,
			validate: func(t *testing.T, resp *proto.UpdateMetricResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Metric)
				assert.Equal(t, "PollCount", resp.Metric.Id)
				assert.Equal(t, proto.MetricType_METRIC_TYPE_COUNTER, resp.Metric.Type)
				assert.Equal(t, int64(30), resp.Metric.GetCounterDelta())
			},
		},
		{
			name:        "nil request",
			request:     nil,
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "nil metric",
			request: &proto.UpdateMetricRequest{
				Metric: nil,
			},
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "empty metric id",
			request: &proto.UpdateMetricRequest{
				Metric: &proto.Metric{
					Id:   "",
					Type: proto.MetricType_METRIC_TYPE_GAUGE,
					Value: &proto.Metric_GaugeValue{
						GaugeValue: 7770.0,
					},
				},
			},
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "unspecified metric type",
			request: &proto.UpdateMetricRequest{
				Metric: &proto.Metric{
					Id:   "Test",
					Type: proto.MetricType_METRIC_TYPE_UNSPECIFIED,
				},
			},
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "gauge without value",
			request: &proto.UpdateMetricRequest{
				Metric: &proto.Metric{
					Id:   "HeapObjects",
					Type: proto.MetricType_METRIC_TYPE_GAUGE,
				},
			},
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "counter without value",
			request: &proto.UpdateMetricRequest{
				Metric: &proto.Metric{
					Id:   "PollCount",
					Type: proto.MetricType_METRIC_TYPE_COUNTER,
				},
			},
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "gauge update error",
			request: &proto.UpdateMetricRequest{
				Metric: &proto.Metric{
					Id:   "HeapObjects",
					Type: proto.MetricType_METRIC_TYPE_GAUGE,
					Value: &proto.Metric_GaugeValue{
						GaugeValue: 7770.0,
					},
				},
			},
			setupMocks: func(storage *MockHandlerStorage, logger *MockGRPCLogger) {
				storage.EXPECT().UpdateGauge(gomock.Any(), "HeapObjects", 7770.0).Return(0.0, errors.New("storage error"))
				logger.EXPECT().Error("Failed to update gauge: %v", gomock.Any())
			},
			expectedErr: codes.Internal,
		},
		{
			name: "counter update error",
			request: &proto.UpdateMetricRequest{
				Metric: &proto.Metric{
					Id:   "PollCount",
					Type: proto.MetricType_METRIC_TYPE_COUNTER,
					Value: &proto.Metric_CounterDelta{
						CounterDelta: 30,
					},
				},
			},
			setupMocks: func(storage *MockHandlerStorage, logger *MockGRPCLogger) {
				storage.EXPECT().UpdateCounter(gomock.Any(), "PollCount", int64(30)).Return(int64(0), errors.New("storage error"))
				logger.EXPECT().Error("Failed to update counter: %v", gomock.Any())
			},
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockHandlerStorage(ctrl)
			mockDB := NewMockDBPinger(ctrl)
			mockLogger := NewMockGRPCLogger(ctrl)

			tt.setupMocks(mockStorage, mockLogger)

			server := NewServer(mockStorage, mockDB, mockLogger)
			resp, err := server.UpdateMetric(context.Background(), tt.request)

			if tt.expectedErr == codes.OK {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
				assert.Nil(t, resp)
			}
		})
	}
}

func TestServer_UpdateMetrics(t *testing.T) {
	tests := []struct {
		name        string
		request     *proto.UpdateMetricsRequest
		setupMocks  func(*MockHandlerStorage, *MockGRPCLogger)
		expectedErr codes.Code
	}{
		{
			name: "successful batch update",
			request: &proto.UpdateMetricsRequest{
				Metrics: []*proto.Metric{
					{
						Id:   "HeapObjects",
						Type: proto.MetricType_METRIC_TYPE_GAUGE,
						Value: &proto.Metric_GaugeValue{
							GaugeValue: 7770.0,
						},
					},
					{
						Id:   "PollCount",
						Type: proto.MetricType_METRIC_TYPE_COUNTER,
						Value: &proto.Metric_CounterDelta{
							CounterDelta: 30,
						},
					},
				},
			},
			setupMocks: func(storage *MockHandlerStorage, logger *MockGRPCLogger) {
				expectedMetrics := model.Metrics{
					{
						ID:    "HeapObjects",
						MType: model.GaugeType,
						Value: func() *float64 { v := 7770.0; return &v }(),
					},
					{
						ID:    "PollCount",
						MType: model.CounterType,
						Delta: func() *int64 { v := int64(30); return &v }(),
					},
				}
				storage.EXPECT().UpdateMetricsBatch(gomock.Any(), expectedMetrics).Return(nil)
			},
			expectedErr: codes.OK,
		},
		{
			name: "empty metrics",
			request: &proto.UpdateMetricsRequest{
				Metrics: []*proto.Metric{},
			},
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.OK,
		},
		{
			name:        "nil request",
			request:     nil,
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "metric with empty id",
			request: &proto.UpdateMetricsRequest{
				Metrics: []*proto.Metric{
					{
						Id:   "",
						Type: proto.MetricType_METRIC_TYPE_GAUGE,
						Value: &proto.Metric_GaugeValue{
							GaugeValue: 7770.0,
						},
					},
				},
			},
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "batch update error",
			request: &proto.UpdateMetricsRequest{
				Metrics: []*proto.Metric{
					{
						Id:   "HeapObjects",
						Type: proto.MetricType_METRIC_TYPE_GAUGE,
						Value: &proto.Metric_GaugeValue{
							GaugeValue: 7770.0,
						},
					},
				},
			},
			setupMocks: func(storage *MockHandlerStorage, logger *MockGRPCLogger) {
				storage.EXPECT().UpdateMetricsBatch(gomock.Any(), gomock.Any()).Return(errors.New("batch update error"))
				logger.EXPECT().Error("Failed to update metrics batch: %v", gomock.Any())
			},
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockHandlerStorage(ctrl)
			mockDB := NewMockDBPinger(ctrl)
			mockLogger := NewMockGRPCLogger(ctrl)

			tt.setupMocks(mockStorage, mockLogger)

			server := NewServer(mockStorage, mockDB, mockLogger)
			resp, err := server.UpdateMetrics(context.Background(), tt.request)

			if tt.expectedErr == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
				assert.Nil(t, resp)
			}
		})
	}
}

func TestServer_GetMetric(t *testing.T) {
	tests := []struct {
		name        string
		request     *proto.GetMetricRequest
		setupMocks  func(*MockHandlerStorage, *MockGRPCLogger)
		expectedErr codes.Code
		validate    func(*testing.T, *proto.GetMetricResponse)
	}{
		{
			name: "successful gauge get",
			request: &proto.GetMetricRequest{
				Id:   "HeapObjects",
				Type: proto.MetricType_METRIC_TYPE_GAUGE,
			},
			setupMocks: func(storage *MockHandlerStorage, logger *MockGRPCLogger) {
				storage.EXPECT().GetGauge(gomock.Any(), "HeapObjects").Return(7770.0, true)
			},
			expectedErr: codes.OK,
			validate: func(t *testing.T, resp *proto.GetMetricResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Metric)
				assert.Equal(t, "HeapObjects", resp.Metric.Id)
				assert.Equal(t, proto.MetricType_METRIC_TYPE_GAUGE, resp.Metric.Type)
				assert.Equal(t, 7770.0, resp.Metric.GetGaugeValue())
			},
		},
		{
			name: "successful counter get",
			request: &proto.GetMetricRequest{
				Id:   "PollCount",
				Type: proto.MetricType_METRIC_TYPE_COUNTER,
			},
			setupMocks: func(storage *MockHandlerStorage, logger *MockGRPCLogger) {
				storage.EXPECT().GetCounter(gomock.Any(), "PollCount").Return(int64(30), true)
			},
			expectedErr: codes.OK,
			validate: func(t *testing.T, resp *proto.GetMetricResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Metric)
				assert.Equal(t, "PollCount", resp.Metric.Id)
				assert.Equal(t, proto.MetricType_METRIC_TYPE_COUNTER, resp.Metric.Type)
				assert.Equal(t, int64(30), resp.Metric.GetCounterDelta())
			},
		},
		{
			name:        "nil request",
			request:     nil,
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "empty id",
			request: &proto.GetMetricRequest{
				Id:   "",
				Type: proto.MetricType_METRIC_TYPE_GAUGE,
			},
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "unspecified type",
			request: &proto.GetMetricRequest{
				Id:   "Test",
				Type: proto.MetricType_METRIC_TYPE_UNSPECIFIED,
			},
			setupMocks:  func(*MockHandlerStorage, *MockGRPCLogger) {},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "gauge not found",
			request: &proto.GetMetricRequest{
				Id:   "NonExistent",
				Type: proto.MetricType_METRIC_TYPE_GAUGE,
			},
			setupMocks: func(storage *MockHandlerStorage, logger *MockGRPCLogger) {
				storage.EXPECT().GetGauge(gomock.Any(), "NonExistent").Return(0.0, false)
			},
			expectedErr: codes.NotFound,
		},
		{
			name: "counter not found",
			request: &proto.GetMetricRequest{
				Id:   "NonExistent",
				Type: proto.MetricType_METRIC_TYPE_COUNTER,
			},
			setupMocks: func(storage *MockHandlerStorage, logger *MockGRPCLogger) {
				storage.EXPECT().GetCounter(gomock.Any(), "NonExistent").Return(int64(0), false)
			},
			expectedErr: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockHandlerStorage(ctrl)
			mockDB := NewMockDBPinger(ctrl)
			mockLogger := NewMockGRPCLogger(ctrl)

			tt.setupMocks(mockStorage, mockLogger)

			server := NewServer(mockStorage, mockDB, mockLogger)
			resp, err := server.GetMetric(context.Background(), tt.request)

			if tt.expectedErr == codes.OK {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
				assert.Nil(t, resp)
			}
		})
	}
}

func TestServer_Ping(t *testing.T) {
	tests := []struct {
		name        string
		request     *proto.PingRequest
		setupMocks  func(*MockDBPinger, *MockGRPCLogger)
		expectedErr codes.Code
	}{
		{
			name:    "successful ping",
			request: &proto.PingRequest{},
			setupMocks: func(db *MockDBPinger, logger *MockGRPCLogger) {
				db.EXPECT().Ping(gomock.Any()).Return(nil)
			},
			expectedErr: codes.OK,
		},
		{
			name:    "ping with db error",
			request: &proto.PingRequest{},
			setupMocks: func(db *MockDBPinger, logger *MockGRPCLogger) {
				db.EXPECT().Ping(gomock.Any()).Return(errors.New("db connection failed"))
				logger.EXPECT().Error("Database connection failed: %v", gomock.Any())
			},
			expectedErr: codes.Unavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockHandlerStorage(ctrl)
			mockDB := NewMockDBPinger(ctrl)
			mockLogger := NewMockGRPCLogger(ctrl)

			tt.setupMocks(mockDB, mockLogger)

			server := NewServer(mockStorage, mockDB, mockLogger)
			resp, err := server.Ping(context.Background(), tt.request)

			if tt.expectedErr == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
				assert.Nil(t, resp)
			}
		})
	}

	t.Run("ping without db", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := NewMockHandlerStorage(ctrl)
		mockLogger := NewMockGRPCLogger(ctrl)

		server := NewServer(mockStorage, nil, mockLogger)
		resp, err := server.Ping(context.Background(), &proto.PingRequest{})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestLoggerInterceptor(t *testing.T) {
	tests := []struct {
		name         string
		handlerResp  interface{}
		handlerErr   error
		expectedCode codes.Code
	}{
		{
			name:         "successful request",
			handlerResp:  &proto.PingResponse{},
			handlerErr:   nil,
			expectedCode: codes.OK,
		},
		{
			name:         "request with error",
			handlerResp:  nil,
			handlerErr:   status.Error(codes.InvalidArgument, "test error"),
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockGRPCLogger(ctrl)

			mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return tt.handlerResp, tt.handlerErr
			}

			interceptor := LoggerInterceptor(mockLogger)

			mockLogger.EXPECT().Info("method=%s status=%s duration=%s", "/test.TestService/TestMethod", tt.expectedCode.String(), gomock.Any()).Times(1)

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.TestService/TestMethod",
			}

			resp, err := interceptor(context.Background(), &proto.PingRequest{}, info, mockHandler)

			assert.Equal(t, tt.handlerResp, resp)
			assert.Equal(t, tt.handlerErr, err)
		})
	}
}
