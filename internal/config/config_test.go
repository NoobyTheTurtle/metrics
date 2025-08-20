package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentDefaultConfig_Success(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "agent_test.json")

	expectedConfig := AgentDefaultConfig{
		ServerAddress:  "localhost:8080",
		LogLevel:       "info",
		AppEnv:         "development",
		PollInterval:   2,
		ReportInterval: 10,
		Key:            "test-key",
		RateLimit:      1,
		CryptoKey:      "/path/to/key.pem",
	}

	configData, err := json.Marshal(expectedConfig)
	require.NoError(t, err)

	err = os.WriteFile(configFile, configData, 0o644)
	require.NoError(t, err)

	config, err := NewAgentDefaultConfig(configFile)

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, expectedConfig.ServerAddress, config.ServerAddress)
	assert.Equal(t, expectedConfig.LogLevel, config.LogLevel)
	assert.Equal(t, expectedConfig.AppEnv, config.AppEnv)
	assert.Equal(t, expectedConfig.PollInterval, config.PollInterval)
	assert.Equal(t, expectedConfig.ReportInterval, config.ReportInterval)
	assert.Equal(t, expectedConfig.Key, config.Key)
	assert.Equal(t, expectedConfig.RateLimit, config.RateLimit)
	assert.Equal(t, expectedConfig.CryptoKey, config.CryptoKey)
}

func TestNewServerDefaultConfig_Success(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "server_test.json")

	expectedConfig := ServerDefaultConfig{
		ServerAddress:   "localhost:8080",
		LogLevel:        "info",
		AppEnv:          "development",
		Key:             "test-key",
		CryptoKey:       "/path/to/key.pem",
		StoreInterval:   300,
		FileStoragePath: "tmp/metrics-db.json",
		Restore:         false,
		DatabaseDSN:     "postgres://user:pass@localhost:5432/metrics",
	}

	configData, err := json.Marshal(expectedConfig)
	require.NoError(t, err)

	err = os.WriteFile(configFile, configData, 0o644)
	require.NoError(t, err)

	config, err := NewServerDefaultConfig(configFile)

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, expectedConfig.ServerAddress, config.ServerAddress)
	assert.Equal(t, expectedConfig.LogLevel, config.LogLevel)
	assert.Equal(t, expectedConfig.AppEnv, config.AppEnv)
	assert.Equal(t, expectedConfig.Key, config.Key)
	assert.Equal(t, expectedConfig.CryptoKey, config.CryptoKey)
	assert.Equal(t, expectedConfig.StoreInterval, config.StoreInterval)
	assert.Equal(t, expectedConfig.FileStoragePath, config.FileStoragePath)
	assert.Equal(t, expectedConfig.Restore, config.Restore)
	assert.Equal(t, expectedConfig.DatabaseDSN, config.DatabaseDSN)
}

func TestNewAgentDefaultConfig_FileNotFound_Error(t *testing.T) {
	nonExistentPath := "/path/that/does/not/exist/config.json"

	config, err := NewAgentDefaultConfig(nonExistentPath)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "error reading config file")
	assert.Contains(t, err.Error(), nonExistentPath)
}

func TestNewServerDefaultConfig_FileNotFound_Error(t *testing.T) {
	nonExistentPath := "/path/that/does/not/exist/config.json"

	config, err := NewServerDefaultConfig(nonExistentPath)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "error reading config file")
	assert.Contains(t, err.Error(), nonExistentPath)
}

func TestNewAgentDefaultConfig_InvalidJSON_Error(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid_agent.json")

	invalidJSON := `{
		"server_address": "localhost:8080",
		"log_level": "info",
		"app_env": "development"
		// missing comma and invalid syntax
		"poll_interval": 2
	}`

	err := os.WriteFile(configFile, []byte(invalidJSON), 0o644)
	require.NoError(t, err)

	config, err := NewAgentDefaultConfig(configFile)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "error parsing config file")
	assert.Contains(t, err.Error(), configFile)
}

func TestNewServerDefaultConfig_InvalidJSON_Error(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid_server.json")

	invalidJSON := `{
		"server_address": "localhost:8080",
		"log_level": "info",
		"app_env": "development",
		"store_interval": "invalid_number"
	}`

	err := os.WriteFile(configFile, []byte(invalidJSON), 0o644)
	require.NoError(t, err)

	config, err := NewServerDefaultConfig(configFile)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "error parsing config file")
	assert.Contains(t, err.Error(), configFile)
}

func TestNewAgentDefaultConfig_EmptyFile_Error(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "empty_agent.json")

	err := os.WriteFile(configFile, []byte(""), 0o644)
	require.NoError(t, err)

	config, err := NewAgentDefaultConfig(configFile)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "error parsing config file")
}

func TestNewServerDefaultConfig_EmptyFile_Error(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "empty_server.json")

	err := os.WriteFile(configFile, []byte(""), 0o644)
	require.NoError(t, err)

	config, err := NewServerDefaultConfig(configFile)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "error parsing config file")
}

func TestNewAgentDefaultConfig_PartialConfig_Success(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "partial_agent.json")

	partialJSON := `{
		"server_address": "test.example.com:9090",
		"log_level": "debug"
	}`

	err := os.WriteFile(configFile, []byte(partialJSON), 0o644)
	require.NoError(t, err)

	config, err := NewAgentDefaultConfig(configFile)

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test.example.com:9090", config.ServerAddress)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, "", config.AppEnv)
	assert.Equal(t, uint(0), config.PollInterval)
}

func TestNewServerDefaultConfig_PartialConfig_Success(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "partial_server.json")

	partialJSON := `{
		"server_address": "test.example.com:9090",
		"log_level": "debug",
		"restore": true
	}`

	err := os.WriteFile(configFile, []byte(partialJSON), 0o644)
	require.NoError(t, err)

	config, err := NewServerDefaultConfig(configFile)

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test.example.com:9090", config.ServerAddress)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, true, config.Restore)
	assert.Equal(t, "", config.AppEnv)
	assert.Equal(t, uint(0), config.StoreInterval)
	assert.Equal(t, "", config.DatabaseDSN)
}
