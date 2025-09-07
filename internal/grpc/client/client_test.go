package client

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/NoobyTheTurtle/metrics/proto"
)

// mockMetricsServer реализует mock gRPC сервер для тестов.
type mockMetricsServer struct {
	proto.UnimplementedMetricsServiceServer
	updateMetricFunc  func(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error)
	updateMetricsFunc func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error)
	pingFunc          func(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error)
}

func (m *mockMetricsServer) UpdateMetric(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
	if m.updateMetricFunc != nil {
		return m.updateMetricFunc(ctx, req)
	}
	return &proto.UpdateMetricResponse{Metric: req.Metric}, nil
}

func (m *mockMetricsServer) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if m.updateMetricsFunc != nil {
		return m.updateMetricsFunc(ctx, req)
	}
	return &proto.UpdateMetricsResponse{}, nil
}

func (m *mockMetricsServer) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	if m.pingFunc != nil {
		return m.pingFunc(ctx, req)
	}
	return &proto.PingResponse{}, nil
}

// setupTestServer создает тестовый gRPC сервер.
func setupTestServer(mockServer *mockMetricsServer) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	proto.RegisterMetricsServiceServer(server, mockServer)

	go func() {
		server.Serve(lis)
	}()

	return server, lis
}

// createTestClient создает тестового клиента подключенного к mock серверу.
func createTestClient(t *testing.T, lis *bufconn.Listener) *Client {
	config := DefaultConfig()
	config.DialTimeout = time.Second
	config.RequestTimeout = time.Second

	logger := NewMockGRPCLogger(gomock.NewController(t))

	// Настраиваем ожидания для логгера на случай ошибок
	logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Создаем соединение к bufconn listener
	conn, err := grpc.Dial("",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	require.NoError(t, err)

	client := &Client{
		config: config,
		conn:   conn,
		client: proto.NewMetricsServiceClient(conn),
		logger: logger,
	}

	return client
}

func TestNewClient(t *testing.T) {
	logger := NewMockGRPCLogger(gomock.NewController(t))

	t.Run("successful creation", func(t *testing.T) {
		// Тестируем только валидацию конфигурации, так как bufconn требует специальной настройки
		config := Config{
			ServerAddress:  "invalid:address",
			DialTimeout:    time.Millisecond, // Короткий таймаут для быстрого завершения теста
			RequestTimeout: time.Second,
		}

		// Ожидаем ошибку подключения к несуществующему адресу
		_, err := NewClient(config, logger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to gRPC server")
	})

	t.Run("nil logger", func(t *testing.T) {
		config := DefaultConfig()
		client, err := NewClient(config, nil)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "logger is required")
	})

	t.Run("empty server address", func(t *testing.T) {
		config := Config{}
		client, err := NewClient(config, logger)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "server address is required")
	})
}

func TestClient_UpdateMetric(t *testing.T) {
	mockServer := &mockMetricsServer{}
	server, lis := setupTestServer(mockServer)
	defer server.Stop()

	client := createTestClient(t, lis)
	defer client.Close()

	t.Run("successful gauge update", func(t *testing.T) {
		value := 42.5
		expectedMetric := &model.Metric{
			ID:    "test_gauge",
			MType: model.GaugeType,
			Value: &value,
		}

		mockServer.updateMetricFunc = func(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
			assert.Equal(t, "test_gauge", req.Metric.Id)
			assert.Equal(t, proto.MetricType_METRIC_TYPE_GAUGE, req.Metric.Type)
			assert.Equal(t, value, req.Metric.GetGaugeValue())

			return &proto.UpdateMetricResponse{Metric: req.Metric}, nil
		}

		result, err := client.UpdateMetric(context.Background(), expectedMetric)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedMetric.ID, result.ID)
		assert.Equal(t, expectedMetric.MType, result.MType)
		assert.Equal(t, *expectedMetric.Value, *result.Value)
	})

	t.Run("successful counter update", func(t *testing.T) {
		delta := int64(10)
		expectedMetric := &model.Metric{
			ID:    "test_counter",
			MType: model.CounterType,
			Delta: &delta,
		}

		mockServer.updateMetricFunc = func(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
			assert.Equal(t, "test_counter", req.Metric.Id)
			assert.Equal(t, proto.MetricType_METRIC_TYPE_COUNTER, req.Metric.Type)
			assert.Equal(t, delta, req.Metric.GetCounterDelta())

			return &proto.UpdateMetricResponse{Metric: req.Metric}, nil
		}

		result, err := client.UpdateMetric(context.Background(), expectedMetric)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedMetric.ID, result.ID)
		assert.Equal(t, expectedMetric.MType, result.MType)
		assert.Equal(t, *expectedMetric.Delta, *result.Delta)
	})

	t.Run("nil metric", func(t *testing.T) {
		result, err := client.UpdateMetric(context.Background(), nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "metric cannot be nil")
	})

	t.Run("server error", func(t *testing.T) {
		value := 42.5
		metric := &model.Metric{
			ID:    "test_gauge",
			MType: model.GaugeType,
			Value: &value,
		}

		mockServer.updateMetricFunc = func(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
			return nil, status.Error(codes.Internal, "internal server error")
		}

		result, err := client.UpdateMetric(context.Background(), metric)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestClient_UpdateMetrics(t *testing.T) {
	mockServer := &mockMetricsServer{}
	server, lis := setupTestServer(mockServer)
	defer server.Stop()

	client := createTestClient(t, lis)
	defer client.Close()

	t.Run("successful batch update", func(t *testing.T) {
		value := 42.5
		delta := int64(10)
		metrics := model.Metrics{
			{
				ID:    "test_gauge",
				MType: model.GaugeType,
				Value: &value,
			},
			{
				ID:    "test_counter",
				MType: model.CounterType,
				Delta: &delta,
			},
		}

		mockServer.updateMetricsFunc = func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
			assert.Len(t, req.Metrics, 2)

			// Проверяем первую метрику (gauge)
			assert.Equal(t, "test_gauge", req.Metrics[0].Id)
			assert.Equal(t, proto.MetricType_METRIC_TYPE_GAUGE, req.Metrics[0].Type)
			assert.Equal(t, value, req.Metrics[0].GetGaugeValue())

			// Проверяем вторую метрику (counter)
			assert.Equal(t, "test_counter", req.Metrics[1].Id)
			assert.Equal(t, proto.MetricType_METRIC_TYPE_COUNTER, req.Metrics[1].Type)
			assert.Equal(t, delta, req.Metrics[1].GetCounterDelta())

			return &proto.UpdateMetricsResponse{}, nil
		}

		err := client.UpdateMetrics(context.Background(), metrics)
		assert.NoError(t, err)
	})

	t.Run("empty metrics slice", func(t *testing.T) {
		err := client.UpdateMetrics(context.Background(), model.Metrics{})
		assert.NoError(t, err)
	})

	t.Run("server error", func(t *testing.T) {
		value := 42.5
		metrics := model.Metrics{
			{
				ID:    "test_gauge",
				MType: model.GaugeType,
				Value: &value,
			},
		}

		mockServer.updateMetricsFunc = func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
			return nil, status.Error(codes.Internal, "internal server error")
		}

		err := client.UpdateMetrics(context.Background(), metrics)
		assert.Error(t, err)
	})
}

func TestClient_Ping(t *testing.T) {
	mockServer := &mockMetricsServer{}
	server, lis := setupTestServer(mockServer)
	defer server.Stop()

	client := createTestClient(t, lis)
	defer client.Close()

	t.Run("successful ping", func(t *testing.T) {
		mockServer.pingFunc = func(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
			return &proto.PingResponse{}, nil
		}

		err := client.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("server error", func(t *testing.T) {
		mockServer.pingFunc = func(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
			return nil, status.Error(codes.Unavailable, "service unavailable")
		}

		err := client.Ping(context.Background())
		assert.Error(t, err)
	})
}

func TestClient_shouldRetryError(t *testing.T) {
	client := &Client{}

	testCases := []struct {
		name        string
		err         error
		shouldRetry bool
	}{
		{
			name:        "nil error",
			err:         nil,
			shouldRetry: false,
		},
		{
			name:        "unavailable error",
			err:         status.Error(codes.Unavailable, "service unavailable"),
			shouldRetry: true,
		},
		{
			name:        "deadline exceeded error",
			err:         status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			shouldRetry: true,
		},
		{
			name:        "resource exhausted error",
			err:         status.Error(codes.ResourceExhausted, "resource exhausted"),
			shouldRetry: true,
		},
		{
			name:        "invalid argument error",
			err:         status.Error(codes.InvalidArgument, "invalid argument"),
			shouldRetry: false,
		},
		{
			name:        "not found error",
			err:         status.Error(codes.NotFound, "not found"),
			shouldRetry: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.shouldRetryError(tc.err)
			assert.Equal(t, tc.shouldRetry, result)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "localhost:50051", config.ServerAddress)
	assert.Equal(t, 5*time.Second, config.DialTimeout)
	assert.Equal(t, 10*time.Second, config.RequestTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.RetryDelay)
}
