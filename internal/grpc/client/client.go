// Package client предоставляет gRPC клиент для отправки метрик на сервер.
package client

import (
	"context"
	"fmt"
	"time"

	grpc_client "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/NoobyTheTurtle/metrics/internal/grpc"
	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/NoobyTheTurtle/metrics/internal/retry"
	"github.com/NoobyTheTurtle/metrics/proto"
)

type Config struct {
	ServerAddress string

	DialTimeout time.Duration

	RequestTimeout time.Duration

	MaxRetries int

	RetryDelay time.Duration
}

func DefaultConfig() Config {
	return Config{
		ServerAddress:  "localhost:50051",
		DialTimeout:    5 * time.Second,
		RequestTimeout: 10 * time.Second,
		MaxRetries:     3,
		RetryDelay:     100 * time.Millisecond,
	}
}

type Client struct {
	config Config
	conn   *grpc_client.ClientConn
	client proto.MetricsServiceClient
	logger GRPCLogger
}

func NewClient(config Config, logger GRPCLogger) (*Client, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if config.ServerAddress == "" {
		return nil, fmt.Errorf("server address is required")
	}

	client := &Client{
		config: config,
		logger: logger,
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return client, nil
}

func (c *Client) connect() error {
	conn, err := grpc_client.NewClient(c.config.ServerAddress,
		grpc_client.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC client for server at %s: %w", c.config.ServerAddress, err)
	}

	c.conn = conn
	c.client = proto.NewMetricsServiceClient(conn)

	return nil
}

func (c *Client) UpdateMetric(ctx context.Context, metric *model.Metric) (*model.Metric, error) {
	if metric == nil {
		return nil, fmt.Errorf("metric cannot be nil")
	}

	protoMetric, err := grpc.ConvertInternalMetricToProto(metric)
	if err != nil {
		return nil, fmt.Errorf("failed to convert metric to protobuf: %w", err)
	}

	req := &proto.UpdateMetricRequest{
		Metric: protoMetric,
	}

	var resp *proto.UpdateMetricResponse
	op := func() error {
		requestCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		defer cancel()

		var opErr error
		resp, opErr = c.client.UpdateMetric(requestCtx, req)
		return opErr
	}

	err = retry.WithRetries(op, c.shouldRetryError)
	if err != nil {
		c.logger.Error("Failed to update metric %s after retries: %v", metric.ID, err)
		return nil, fmt.Errorf("failed to update metric %s: %w", metric.ID, err)
	}

	resultMetric, err := grpc.ConvertProtoMetricToInternal(resp.Metric)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response metric: %w", err)
	}

	return resultMetric, nil
}

func (c *Client) UpdateMetrics(ctx context.Context, metrics model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	protoMetrics, err := grpc.ConvertInternalMetricsToProto(metrics)
	if err != nil {
		return fmt.Errorf("failed to convert metrics to protobuf: %w", err)
	}

	req := &proto.UpdateMetricsRequest{
		Metrics: protoMetrics,
	}

	op := func() error {
		requestCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		defer cancel()

		_, opErr := c.client.UpdateMetrics(requestCtx, req)
		return opErr
	}

	err = retry.WithRetries(op, c.shouldRetryError)
	if err != nil {
		c.logger.Error("Failed to update metrics batch after retries: %v", err)
		return fmt.Errorf("failed to update metrics batch: %w", err)
	}

	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	req := &proto.PingRequest{}

	op := func() error {
		requestCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		defer cancel()

		_, opErr := c.client.Ping(requestCtx, req)
		return opErr
	}

	err := retry.WithRetries(op, c.shouldRetryError)
	if err != nil {
		c.logger.Error("Ping failed after retries: %v", err)
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			c.logger.Error("Failed to close gRPC connection: %v", err)
			return fmt.Errorf("failed to close gRPC connection: %w", err)
		}
	}
	return nil
}

func (c *Client) shouldRetryError(err error) bool {
	if err == nil {
		return false
	}

	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
			return true
		case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.PermissionDenied:
			return false
		default:
			return true
		}
	}

	return retry.RequestErrorChecker(err)
}
