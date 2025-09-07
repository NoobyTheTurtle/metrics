package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NoobyTheTurtle/metrics/internal/testutil"
)

func TestNewServerConfig(t *testing.T) {
	oldArgs := os.Args
	oldEnv := map[string]string{}
	for _, env := range []string{"ADDRESS", "LOG_LEVEL", "APP_ENV", "DATABASE_DSN", "TRUSTED_SUBNET"} {
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
		expected       *ServerConfig
		expectedErrMsg string
	}{
		{
			name: "default values",
			args: []string{"test"},
			expected: &ServerConfig{
				ConfigPath:    "../../configs/server.json",
				ServerAddress: "localhost:8080",
				LogLevel:      "info",
				AppEnv:        "development",
				DatabaseDSN:   "",
				TrustedSubnet: "",
			},
		},
		{
			name: "command line arguments",
			args: []string{"test", "-a", "localhost:9090"},
			expected: &ServerConfig{
				ConfigPath:    "../../configs/server.json",
				ServerAddress: "localhost:9090",
				LogLevel:      "info",
				AppEnv:        "development",
				DatabaseDSN:   "",
				TrustedSubnet: "",
			},
		},
		{
			name: "database dsn command line argument",
			args: []string{"test", "-d", "postgres://testuser:testpass@testhost:5432/testdb?sslmode=disable"},
			expected: &ServerConfig{
				ConfigPath:    "../../configs/server.json",
				ServerAddress: "localhost:8080",
				LogLevel:      "info",
				AppEnv:        "development",
				DatabaseDSN:   "postgres://testuser:testpass@testhost:5432/testdb?sslmode=disable",
				TrustedSubnet: "",
			},
		},
		{
			name: "trusted subnet command line argument",
			args: []string{"test", "-t", "172.16.0.0/16"},
			expected: &ServerConfig{
				ConfigPath:    "../../configs/server.json",
				ServerAddress: "localhost:8080",
				LogLevel:      "info",
				AppEnv:        "development",
				DatabaseDSN:   "",
				TrustedSubnet: "172.16.0.0/16",
			},
		},
		{
			name: "environment variables",
			args: []string{"test"},
			envs: map[string]string{
				"ADDRESS":        "localhost:7070",
				"LOG_LEVEL":      "debug",
				"APP_ENV":        "test",
				"DATABASE_DSN":   "postgres://envuser:envpass@envhost:5432/envdb?sslmode=disable",
				"TRUSTED_SUBNET": "192.168.1.0/24",
			},
			expected: &ServerConfig{
				ConfigPath:    "../../configs/server.json",
				ServerAddress: "localhost:7070",
				LogLevel:      "debug",
				AppEnv:        "test",
				DatabaseDSN:   "postgres://envuser:envpass@envhost:5432/envdb?sslmode=disable",
				TrustedSubnet: "192.168.1.0/24",
			},
		},
		{
			name: "environment variables override flags",
			args: []string{"test", "-a", "localhost:9090", "-d", "postgres://flaguser:flagpass@flaghost:5432/flagdb?sslmode=disable"},
			envs: map[string]string{
				"ADDRESS":        "localhost:7070",
				"LOG_LEVEL":      "debug",
				"APP_ENV":        "test",
				"DATABASE_DSN":   "postgres://envuser:envpass@envhost:5432/envdb?sslmode=disable",
				"TRUSTED_SUBNET": "10.0.0.0/8",
			},
			expected: &ServerConfig{
				ConfigPath:    "../../configs/server.json",
				ServerAddress: "localhost:7070",
				LogLevel:      "debug",
				AppEnv:        "test",
				DatabaseDSN:   "postgres://envuser:envpass@envhost:5432/envdb?sslmode=disable",
				TrustedSubnet: "10.0.0.0/8",
			},
		},
		{
			name:           "unknown arguments",
			args:           []string{"test", "unknown"},
			expectedErrMsg: "config.ServerConfig.parseFlags: unknown command line arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ResetFlags()
			// Set args with custom config path for tests
			args := append(tt.args, "-c", "../../configs/server.json")
			os.Args = args

			for k, v := range tt.envs {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envs {
					os.Unsetenv(k)
				}
			}()

			config, err := NewServerConfig()

			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ConfigPath, config.ConfigPath)
				assert.Equal(t, tt.expected.ServerAddress, config.ServerAddress)
				assert.Equal(t, tt.expected.LogLevel, config.LogLevel)
				assert.Equal(t, tt.expected.AppEnv, config.AppEnv)
				assert.Equal(t, tt.expected.DatabaseDSN, config.DatabaseDSN)
				assert.Equal(t, tt.expected.TrustedSubnet, config.TrustedSubnet)
			}
		})
	}
}
