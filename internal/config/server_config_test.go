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
	for _, env := range []string{"ADDRESS", "LOG_LEVEL", "APP_ENV", "DATABASE_DSN"} {
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
				ServerAddress: "localhost:8080",
				LogLevel:      "info",
				AppEnv:        "development",
				DatabaseDSN:   "host=localhost user=postgres password=postgres dbname=metrics port=5432 sslmode=disable",
			},
		},
		{
			name: "command line arguments",
			args: []string{"test", "-a", "localhost:9090"},
			expected: &ServerConfig{
				ServerAddress: "localhost:9090",
				LogLevel:      "info",
				AppEnv:        "development",
				DatabaseDSN:   "host=localhost user=postgres password=postgres dbname=metrics port=5432 sslmode=disable",
			},
		},
		{
			name: "database dsn command line argument",
			args: []string{"test", "-d", "host=testhost user=testuser password=testpass dbname=testdb port=5432 sslmode=disable"},
			expected: &ServerConfig{
				ServerAddress: "localhost:8080",
				LogLevel:      "info",
				AppEnv:        "development",
				DatabaseDSN:   "host=testhost user=testuser password=testpass dbname=testdb port=5432 sslmode=disable",
			},
		},
		{
			name: "environment variables",
			args: []string{"test"},
			envs: map[string]string{
				"ADDRESS":      "localhost:7070",
				"LOG_LEVEL":    "debug",
				"APP_ENV":      "test",
				"DATABASE_DSN": "host=envhost user=envuser password=envpass dbname=envdb port=5432 sslmode=disable",
			},
			expected: &ServerConfig{
				ServerAddress: "localhost:7070",
				LogLevel:      "debug",
				AppEnv:        "test",
				DatabaseDSN:   "host=envhost user=envuser password=envpass dbname=envdb port=5432 sslmode=disable",
			},
		},
		{
			name: "environment variables override flags",
			args: []string{"test", "-a", "localhost:9090", "-d", "host=flaghost user=flaguser password=flagpass dbname=flagdb port=5432 sslmode=disable"},
			envs: map[string]string{
				"ADDRESS":      "localhost:7070",
				"LOG_LEVEL":    "debug",
				"APP_ENV":      "test",
				"DATABASE_DSN": "host=envhost user=envuser password=envpass dbname=envdb port=5432 sslmode=disable",
			},
			expected: &ServerConfig{
				ServerAddress: "localhost:7070",
				LogLevel:      "debug",
				AppEnv:        "test",
				DatabaseDSN:   "host=envhost user=envuser password=envpass dbname=envdb port=5432 sslmode=disable",
			},
		},
		{
			name:           "unknown arguments",
			args:           []string{"test", "unknown"},
			expectedErrMsg: "unknown command line arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ResetFlags()
			os.Args = tt.args

			for k, v := range tt.envs {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envs {
					os.Unsetenv(k)
				}
			}()

			config, err := NewServerConfig("../../configs/default.yml")

			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ServerAddress, config.ServerAddress)
				assert.Equal(t, tt.expected.LogLevel, config.LogLevel)
				assert.Equal(t, tt.expected.AppEnv, config.AppEnv)
				assert.Equal(t, tt.expected.DatabaseDSN, config.DatabaseDSN)
			}
		})
	}
}
