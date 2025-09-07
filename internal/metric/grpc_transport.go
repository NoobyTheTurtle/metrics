package metric

import (
	"context"
	"fmt"
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/grpc/client"
	"github.com/NoobyTheTurtle/metrics/internal/model"
)

type GRPCTransport struct {
	client client.MetricsClient
	logger MetricsLogger
}

func NewGRPCTransport(serverAddress string, logger MetricsLogger) (*GRPCTransport, error) {
	config := client.Config{
		ServerAddress:  serverAddress,
		DialTimeout:    5 * time.Second,
		RequestTimeout: 10 * time.Second,
		MaxRetries:     3,
		RetryDelay:     100 * time.Millisecond,
	}

	grpcClient, err := client.NewClient(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &GRPCTransport{
		client: grpcClient,
		logger: logger,
	}, nil
}

func (gt *GRPCTransport) SendMetrics(ctx context.Context, metrics model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	err := gt.client.UpdateMetrics(ctx, metrics)
	if err != nil {
		gt.logger.Warn("Failed to send metrics via gRPC: %v", err)
		return err
	}

	return nil
}

func (gt *GRPCTransport) Close() error {
	if gt.client == nil {
		return nil
	}
	return gt.client.Close()
}
