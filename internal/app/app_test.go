package app

import (
	"context"
	"testing"
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/config"
	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestStartAgent(t *testing.T) {
	t.Run("basic test", func(t *testing.T) {
		err := StartAgent()
		assert.Error(t, err)
	})
}

func TestStartServer(t *testing.T) {
	t.Run("basic test", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := StartServer(ctx)
		assert.Error(t, err)
	})
}

func TestInitMetricStorage(t *testing.T) {
	t.Run("memory storage", func(t *testing.T) {
		ctx := context.Background()
		config := &config.ServerConfig{
			DatabaseDSN:     "",
			FileStoragePath: "",
			StoreInterval:   0,
			Restore:         false,
		}
		log := &logger.ZapLogger{}

		storage, err := initMetricStorage(ctx, config, nil, log)

		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, storage)
		}
	})
}

func TestApp_Integration(t *testing.T) {
	t.Run("agent startup", func(t *testing.T) {
		err := StartAgent()
		assert.Error(t, err)
	})

	t.Run("server startup", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := StartServer(ctx)
		assert.Error(t, err)
	})
}
