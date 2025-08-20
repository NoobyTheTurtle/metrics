package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewZapLogger(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		isDev   bool
		wantErr bool
	}{
		{
			name:    "development logger with debug level",
			level:   "debug",
			isDev:   true,
			wantErr: false,
		},
		{
			name:    "production logger with info level",
			level:   "info",
			isDev:   false,
			wantErr: false,
		},
		{
			name:    "production logger with warn level",
			level:   "warn",
			isDev:   false,
			wantErr: false,
		},
		{
			name:    "production logger with error level",
			level:   "error",
			isDev:   false,
			wantErr: false,
		},
		{
			name:    "invalid level",
			level:   "invalid",
			isDev:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewZapLogger(tt.level, tt.isDev)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				assert.NotNil(t, logger.logger)
			}
		})
	}
}

func TestZapLogger_Methods(t *testing.T) {
	logger, err := NewZapLogger("debug", true)
	require.NoError(t, err)
	require.NotNil(t, logger)

	t.Run("Debug", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Debug("debug message")
			logger.Debug("debug message with args: %s", "test")
		})
	})

	t.Run("Info", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Info("info message")
			logger.Info("info message with args: %s", "test")
		})
	})
}

func TestZapLogger_Warn_Success(t *testing.T) {
	logger, err := NewZapLogger("debug", true)
	require.NoError(t, err)
	require.NotNil(t, logger)

	assert.NotPanics(t, func() {
		logger.Warn("warn message")
		logger.Warn("warn message with args: %s, %d", "test", 123)
	})
}

func TestZapLogger_Error_Success(t *testing.T) {
	logger, err := NewZapLogger("debug", true)
	require.NoError(t, err)
	require.NotNil(t, logger)

	assert.NotPanics(t, func() {
		logger.Error("error message")
		logger.Error("error message with args: %s, %d", "test", 456)
	})
}

func TestZapLogger_Fatal_Success(t *testing.T) {
	logger, err := NewZapLogger("debug", true)
	require.NoError(t, err)
	require.NotNil(t, logger)

	assert.NotNil(t, logger.Fatal)
}

func TestZapLogger_Sync_Success(t *testing.T) {
	logger, err := NewZapLogger("debug", true)
	require.NoError(t, err)
	require.NotNil(t, logger)

	syncErr := logger.Sync()
	_ = syncErr
}

func TestZapLogger_Integration(t *testing.T) {
	logger, err := NewZapLogger("info", false)
	require.NoError(t, err)
	require.NotNil(t, logger)

	assert.NotPanics(t, func() {
		logger.Info("test info message")
	})
}
