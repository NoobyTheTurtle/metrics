package metric

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/NoobyTheTurtle/metrics/proto"
)

type mockTransportServer struct {
	proto.UnimplementedMetricsServiceServer
	updateMetricsFunc func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error)
}

func (m *mockTransportServer) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if m.updateMetricsFunc != nil {
		return m.updateMetricsFunc(ctx, req)
	}
	return &proto.UpdateMetricsResponse{}, nil
}

func setupTransportTestServer(mockServer *mockTransportServer) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	proto.RegisterMetricsServiceServer(server, mockServer)

	go func() {
		server.Serve(lis)
	}()

	return server, lis
}

func TestNewGRPCTransport(t *testing.T) {
	logger := NewMockMetricsLogger(gomock.NewController(t))

	t.Run("successful creation with real server", func(t *testing.T) {
		lis, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)
		defer lis.Close()

		server := grpc.NewServer()
		mockServer := &mockTransportServer{}
		proto.RegisterMetricsServiceServer(server, mockServer)

		go func() {
			server.Serve(lis)
		}()
		defer server.GracefulStop()

		time.Sleep(50 * time.Millisecond)

		serverAddr := lis.Addr().String()
		transport, err := NewGRPCTransport(serverAddr, logger)

		assert.NoError(t, err)
		assert.NotNil(t, transport)
		assert.NotNil(t, transport.client)
		assert.Equal(t, logger, transport.logger)

		err = transport.Close()
		assert.NoError(t, err)
	})

	t.Run("creation with invalid server address", func(t *testing.T) {
		transport, err := NewGRPCTransport("invalid:address", logger)
		assert.Error(t, err)
		assert.Nil(t, transport)
		assert.Contains(t, err.Error(), "failed to create gRPC client")
	})

	t.Run("creation with nil logger should fail", func(t *testing.T) {
		transport, err := NewGRPCTransport("localhost:50051", nil)
		assert.Error(t, err)
		assert.Nil(t, transport)
	})
}

func TestGRPCTransport_SendMetrics(t *testing.T) {
	tests := []struct {
		name           string
		metrics        model.Metrics
		setupMockFunc  func(*mockTransportServer)
		expectedError  bool
		errorSubstring string
	}{
		{
			name:    "empty metrics slice",
			metrics: model.Metrics{},
			setupMockFunc: func(m *mockTransportServer) {
			},
			expectedError: false,
		},
		{
			name: "successful metrics send",
			metrics: model.Metrics{
				{
					ID:    "test_gauge",
					MType: model.GaugeType,
					Value: func() *float64 { v := 123.45; return &v }(),
				},
				{
					ID:    "test_counter",
					MType: model.CounterType,
					Delta: func() *int64 { v := int64(10); return &v }(),
				},
			},
			setupMockFunc: func(m *mockTransportServer) {
				m.updateMetricsFunc = func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
					assert.Len(t, req.Metrics, 2)

					assert.Equal(t, "test_gauge", req.Metrics[0].Id)
					assert.Equal(t, proto.MetricType_METRIC_TYPE_GAUGE, req.Metrics[0].Type)
					assert.Equal(t, 123.45, req.Metrics[0].GetGaugeValue())

					assert.Equal(t, "test_counter", req.Metrics[1].Id)
					assert.Equal(t, proto.MetricType_METRIC_TYPE_COUNTER, req.Metrics[1].Type)
					assert.Equal(t, int64(10), req.Metrics[1].GetCounterDelta())

					return &proto.UpdateMetricsResponse{}, nil
				}
			},
			expectedError: false,
		},
		{
			name: "server error",
			metrics: model.Metrics{
				{
					ID:    "test_gauge",
					MType: model.GaugeType,
					Value: func() *float64 { v := 123.45; return &v }(),
				},
			},
			setupMockFunc: func(m *mockTransportServer) {
				m.updateMetricsFunc = func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
					return nil, status.Error(codes.Internal, "internal server error")
				}
			},
			expectedError:  true,
			errorSubstring: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := &mockTransportServer{}
			tt.setupMockFunc(mockServer)

			server, lis := setupTransportTestServer(mockServer)
			defer server.Stop()

			logger := NewMockMetricsLogger(gomock.NewController(t))

			// Set up logger expectations
			if tt.expectedError {
				logger.EXPECT().Warn(gomock.Any(), gomock.Any()).Times(1)
			} else {
				logger.EXPECT().Warn(gomock.Any(), gomock.Any()).Times(0)
			}

			conn, err := grpc.Dial("",
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
					return lis.Dial()
				}),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			require.NoError(t, err)
			defer conn.Close()

			transport := &GRPCTransport{
				client: &testGRPCClient{conn: conn},
				logger: logger,
			}

			ctx := context.Background()
			err = transport.SendMetrics(ctx, tt.metrics)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type testGRPCClient struct {
	conn *grpc.ClientConn
}

func (c *testGRPCClient) UpdateMetric(ctx context.Context, metric *model.Metric) (*model.Metric, error) {
	client := proto.NewMetricsServiceClient(c.conn)

	var protoMetric *proto.Metric
	switch metric.MType {
	case model.GaugeType:
		if metric.Value != nil {
			protoMetric = &proto.Metric{
				Id:   metric.ID,
				Type: proto.MetricType_METRIC_TYPE_GAUGE,
				Value: &proto.Metric_GaugeValue{
					GaugeValue: *metric.Value,
				},
			}
		}
	case model.CounterType:
		if metric.Delta != nil {
			protoMetric = &proto.Metric{
				Id:   metric.ID,
				Type: proto.MetricType_METRIC_TYPE_COUNTER,
				Value: &proto.Metric_CounterDelta{
					CounterDelta: *metric.Delta,
				},
			}
		}
	}

	if protoMetric == nil {
		return nil, assert.AnError
	}

	resp, err := client.UpdateMetric(ctx, &proto.UpdateMetricRequest{
		Metric: protoMetric,
	})
	if err != nil {
		return nil, err
	}

	result := &model.Metric{
		ID:    resp.Metric.Id,
		MType: metric.MType,
	}

	switch resp.Metric.Type {
	case proto.MetricType_METRIC_TYPE_GAUGE:
		value := resp.Metric.GetGaugeValue()
		result.Value = &value
	case proto.MetricType_METRIC_TYPE_COUNTER:
		delta := resp.Metric.GetCounterDelta()
		result.Delta = &delta
	}

	return result, nil
}

func (c *testGRPCClient) UpdateMetrics(ctx context.Context, metrics model.Metrics) error {
	client := proto.NewMetricsServiceClient(c.conn)

	var protoMetrics []*proto.Metric
	for _, metric := range metrics {
		var protoMetric *proto.Metric

		switch metric.MType {
		case model.GaugeType:
			if metric.Value != nil {
				protoMetric = &proto.Metric{
					Id:   metric.ID,
					Type: proto.MetricType_METRIC_TYPE_GAUGE,
					Value: &proto.Metric_GaugeValue{
						GaugeValue: *metric.Value,
					},
				}
			}
		case model.CounterType:
			if metric.Delta != nil {
				protoMetric = &proto.Metric{
					Id:   metric.ID,
					Type: proto.MetricType_METRIC_TYPE_COUNTER,
					Value: &proto.Metric_CounterDelta{
						CounterDelta: *metric.Delta,
					},
				}
			}
		}

		if protoMetric != nil {
			protoMetrics = append(protoMetrics, protoMetric)
		}
	}

	_, err := client.UpdateMetrics(ctx, &proto.UpdateMetricsRequest{
		Metrics: protoMetrics,
	})
	return err
}

func (c *testGRPCClient) Ping(ctx context.Context) error {
	client := proto.NewMetricsServiceClient(c.conn)
	_, err := client.Ping(ctx, &proto.PingRequest{})
	return err
}

func (c *testGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func TestGRPCTransport_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		conn, err := grpc.Dial("localhost:0",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(100*time.Millisecond),
		)

		transport := &GRPCTransport{
			client: &testGRPCClient{conn: conn},
			logger: NewMockMetricsLogger(gomock.NewController(t)),
		}

		err = transport.Close()
		assert.NoError(t, err)
	})

	t.Run("close with nil client", func(t *testing.T) {
		transport := &GRPCTransport{
			client: nil,
			logger: NewMockMetricsLogger(gomock.NewController(t)),
		}

		assert.NotPanics(t, func() {
			transport.Close()
		})
	})
}

func TestGRPCTransport_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer lis.Close()

	server := grpc.NewServer()
	mockServer := &mockTransportServer{
		updateMetricsFunc: func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
			return &proto.UpdateMetricsResponse{}, nil
		},
	}
	proto.RegisterMetricsServiceServer(server, mockServer)

	go func() {
		server.Serve(lis)
	}()
	defer server.GracefulStop()

	time.Sleep(50 * time.Millisecond)

	logger := NewMockMetricsLogger(gomock.NewController(t))
	serverAddr := lis.Addr().String()

	transport, err := NewGRPCTransport(serverAddr, logger)
	require.NoError(t, err)
	defer transport.Close()

	metrics := model.Metrics{
		{
			ID:    "integration_gauge",
			MType: model.GaugeType,
			Value: func() *float64 { v := 99.99; return &v }(),
		},
		{
			ID:    "integration_counter",
			MType: model.CounterType,
			Delta: func() *int64 { v := int64(42); return &v }(),
		},
	}

	ctx := context.Background()
	err = transport.SendMetrics(ctx, metrics)
	assert.NoError(t, err)

	err = transport.SendMetrics(ctx, model.Metrics{})
	assert.NoError(t, err)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = transport.SendMetrics(ctxWithTimeout, metrics)
	assert.NoError(t, err)
}
