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
	for _, env := range []string{"ADDRESS", "POLL_INTERVAL", "REPORT_INTERVAL", "LOG_LEVEL", "APP_ENV"} {
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
				ServerAddress:  "localhost:8080",
				PollInterval:   2,
				ReportInterval: 10,
				LogLevel:       "info",
				AppEnv:         "development",
			},
		},
		{
			name: "command line arguments",
			args: []string{"test", "-a", "localhost:9090", "-p", "5", "-r", "20"},
			expected: &AgentConfig{
				ServerAddress:  "localhost:9090",
				PollInterval:   5,
				ReportInterval: 20,
				LogLevel:       "info",
				AppEnv:         "development",
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
			},
			expected: &AgentConfig{
				ServerAddress:  "localhost:7070",
				PollInterval:   3,
				ReportInterval: 15,
				LogLevel:       "debug",
				AppEnv:         "test",
			},
		},
		{
			name: "environment variables override flags",
			args: []string{"test", "-a", "localhost:9090", "-p", "5", "-r", "20"},
			envs: map[string]string{
				"ADDRESS":         "localhost:7070",
				"POLL_INTERVAL":   "3",
				"REPORT_INTERVAL": "15",
				"LOG_LEVEL":       "debug",
				"APP_ENV":         "test",
			},
			expected: &AgentConfig{
				ServerAddress:  "localhost:7070",
				PollInterval:   3,
				ReportInterval: 15,
				LogLevel:       "debug",
				AppEnv:         "test",
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
			name:           "unknown arguments",
			args:           []string{"test", "unknown"},
			expectedErrMsg: "config.AgentConfig.parseFlags: unknown command line arguments",
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

			config, err := NewAgentConfig("../../configs/default.yml")

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
			}
		})
	}
}
