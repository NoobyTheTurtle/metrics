package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NoobyTheTurtle/metrics/internal/testutil"
)

func TestNewAgentConfig(t *testing.T) {
	oldArgs := os.Args
	oldEnv := map[string]string{}
	for _, env := range []string{"ADDRESS", "POLL_INTERVAL", "REPORT_INTERVAL", "LOG_LEVEL", "APP_ENV", "KEY", "RATE_LIMIT", "CRYPTO_KEY", "GRPC_ADDRESS", "ENABLE_GRPC"} {
		oldEnv[env] = os.Getenv(env)
	}

	defer func() {
		os.Args = oldArgs
		for env, val := range oldEnv {
			if val == "" {
				os.Unsetenv(env)
			} else {
				os.Setenv(env, val)
			}
		}
	}()

	for env := range oldEnv {
		os.Unsetenv(env)
	}

	tests := []struct {
		name           string
		args           []string
		envs           map[string]string
		expected       *AgentConfig
		expectedErrMsg string
	}{
		{
			name: "default values",
			args: []string{"test"},
			expected: &AgentConfig{
				ServerAddress:     "localhost:8080",
				PollInterval:      2,
				ReportInterval:    10,
				LogLevel:          "info",
				AppEnv:            "development",
				Key:               "",
				RateLimit:         1,
				CryptoKey:         "",
				GRPCServerAddress: "localhost:9090",
				EnableGRPC:        false,
			},
		},
		{
			name: "command line arguments",
			args: []string{"test", "-a", "localhost:9090", "-p", "5", "-r", "20", "-grpc-address", "localhost:9091", "-enable-grpc"},
			expected: &AgentConfig{
				ServerAddress:     "localhost:9090",
				PollInterval:      5,
				ReportInterval:    20,
				LogLevel:          "info",
				AppEnv:            "development",
				Key:               "",
				RateLimit:         1,
				CryptoKey:         "",
				GRPCServerAddress: "localhost:9091",
				EnableGRPC:        true,
			},
		},
		{
			name: "environment variables",
			args: []string{"test"},
			envs: map[string]string{
				"ADDRESS":         "localhost:7070",
				"POLL_INTERVAL":   "3",
				"REPORT_INTERVAL": "15",
				"LOG_LEVEL":       "debug",
				"APP_ENV":         "test",
				"KEY":             "test-key",
				"RATE_LIMIT":      "5",
				"CRYPTO_KEY":      "/path/to/key",
				"GRPC_ADDRESS":    "localhost:9092",
				"ENABLE_GRPC":     "true",
			},
			expected: &AgentConfig{
				ServerAddress:     "localhost:7070",
				PollInterval:      3,
				ReportInterval:    15,
				LogLevel:          "debug",
				AppEnv:            "test",
				Key:               "test-key",
				RateLimit:         5,
				CryptoKey:         "/path/to/key",
				GRPCServerAddress: "localhost:9092",
				EnableGRPC:        true,
			},
		},
		{
			name: "environment variables override flags",
			args: []string{"test", "-a", "localhost:9090", "-p", "5", "-r", "20", "-grpc-address", "localhost:9091", "-enable-grpc"},
			envs: map[string]string{
				"ADDRESS":         "localhost:7070",
				"POLL_INTERVAL":   "3",
				"REPORT_INTERVAL": "15",
				"LOG_LEVEL":       "debug",
				"APP_ENV":         "test",
				"KEY":             "env-key",
				"RATE_LIMIT":      "10",
				"CRYPTO_KEY":      "/env/path/to/key",
				"GRPC_ADDRESS":    "localhost:9093",
				"ENABLE_GRPC":     "false",
			},
			expected: &AgentConfig{
				ServerAddress:     "localhost:7070",
				PollInterval:      3,
				ReportInterval:    15,
				LogLevel:          "debug",
				AppEnv:            "test",
				Key:               "env-key",
				RateLimit:         10,
				CryptoKey:         "/env/path/to/key",
				GRPCServerAddress: "localhost:9093",
				EnableGRPC:        false,
			},
		},
		{
			name: "invalid poll interval environment variable",
			args: []string{"test"},
			envs: map[string]string{
				"POLL_INTERVAL": "invalid",
			},
			expectedErrMsg: "config.NewAgentConfig: parsing environment variables",
		},
		{
			name: "invalid report interval environment variable",
			args: []string{"test"},
			envs: map[string]string{
				"REPORT_INTERVAL": "invalid",
			},
			expectedErrMsg: "config.NewAgentConfig: parsing environment variables",
		},
		{
			name: "invalid rate limit environment variable",
			args: []string{"test"},
			envs: map[string]string{
				"RATE_LIMIT": "invalid",
			},
			expectedErrMsg: "config.NewAgentConfig: parsing environment variables",
		},
		{
			name: "invalid enable grpc environment variable",
			args: []string{"test"},
			envs: map[string]string{
				"ENABLE_GRPC": "invalid",
			},
			expectedErrMsg: "config.NewAgentConfig: parsing environment variables",
		},
		{
			name:           "unknown arguments",
			args:           []string{"test", "unknown"},
			expectedErrMsg: "config.AgentConfig.parseFlags: unknown command line arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ResetFlags()

			args := append(tt.args, "-c", "../../configs/agent.json")
			os.Args = args

			for k, v := range tt.envs {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envs {
					os.Unsetenv(k)
				}
			}()

			config, err := NewAgentConfig()

			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ServerAddress, config.ServerAddress)
				assert.Equal(t, tt.expected.PollInterval, config.PollInterval)
				assert.Equal(t, tt.expected.ReportInterval, config.ReportInterval)
				assert.Equal(t, tt.expected.LogLevel, config.LogLevel)
				assert.Equal(t, tt.expected.AppEnv, config.AppEnv)
				assert.Equal(t, tt.expected.Key, config.Key)
				assert.Equal(t, tt.expected.RateLimit, config.RateLimit)
				assert.Equal(t, tt.expected.CryptoKey, config.CryptoKey)
				assert.Equal(t, tt.expected.GRPCServerAddress, config.GRPCServerAddress)
				assert.Equal(t, tt.expected.EnableGRPC, config.EnableGRPC)
			}
		})
	}
}
