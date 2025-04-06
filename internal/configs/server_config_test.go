package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerConfig(t *testing.T) {
	oldArgs := os.Args
	oldEnv := os.Getenv("ADDRESS")

	defer func() {
		os.Args = oldArgs
		if oldEnv == "" {
			os.Unsetenv("ADDRESS")
		} else {
			os.Setenv("ADDRESS", oldEnv)
		}
	}()

	os.Unsetenv("ADDRESS")

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
			},
		},
		{
			name: "command line arguments",
			args: []string{"test", "-a", "localhost:9090"},
			expected: &ServerConfig{
				ServerAddress: "localhost:9090",
			},
		},
		{
			name: "environment variables",
			args: []string{"test"},
			envs: map[string]string{
				"ADDRESS": "localhost:7070",
			},
			expected: &ServerConfig{
				ServerAddress: "localhost:7070",
			},
		},
		{
			name: "environment variables override flags",
			args: []string{"test", "-a", "localhost:9090"},
			envs: map[string]string{
				"ADDRESS": "localhost:7070",
			},
			expected: &ServerConfig{
				ServerAddress: "localhost:7070",
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
			resetFlags()
			os.Args = tt.args

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
				assert.Equal(t, tt.expected.ServerAddress, config.ServerAddress)
			}
		})
	}
}
