package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type AgentDefaultConfig struct {
	ServerAddress  string `json:"server_address"`
	LogLevel       string `json:"log_level"`
	AppEnv         string `json:"app_env"`
	PollInterval   uint   `json:"poll_interval"`
	ReportInterval uint   `json:"report_interval"`
	Key            string `json:"key"`
	RateLimit      uint   `json:"rate_limit"`
	CryptoKey      string `json:"crypto_key"`
}

type ServerDefaultConfig struct {
	ServerAddress   string `json:"server_address"`
	LogLevel        string `json:"log_level"`
	AppEnv          string `json:"app_env"`
	Key             string `json:"key"`
	CryptoKey       string `json:"crypto_key"`
	StoreInterval   uint   `json:"store_interval"`
	FileStoragePath string `json:"file_storage_path"`
	Restore         bool   `json:"restore"`
	DatabaseDSN     string `json:"database_dsn"`
	TrustedSubnet   string `json:"trusted_subnet"`
}

func NewAgentDefaultConfig(configPath string) (*AgentDefaultConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config.NewAgentDefaultConfig: error reading config file '%s': %w", configPath, err)
	}

	var config AgentDefaultConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("config.NewAgentDefaultConfig: error parsing config file '%s': %w", configPath, err)
	}

	return &config, nil
}

func NewServerDefaultConfig(configPath string) (*ServerDefaultConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config.NewServerDefaultConfig: error reading config file '%s': %w", configPath, err)
	}

	var config ServerDefaultConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("config.NewServerDefaultConfig: error parsing config file '%s': %w", configPath, err)
	}

	return &config, nil
}
