package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DefaultConfig struct {
	ServerAddress string `yaml:"server_address"`
	LogLevel      string `yaml:"log_level"`
	AppEnv        string `yaml:"app_env"`

	PollInterval   uint `yaml:"poll_interval"`
	ReportInterval uint `yaml:"report_interval"`

	StoreInterval   uint   `yaml:"store_interval"`
	FileStoragePath string `yaml:"file_storage_path"`
	Restore         bool   `yaml:"restore"`

	DatabaseDSN string `yaml:"database_dsn"`

	Key       string `yaml:"key"`
	RateLimit uint   `yaml:"rate_limit"`
}

func NewDefaultConfig(configPath string) (*DefaultConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config.NewDefaultConfig: error reading config file '%s': %w", configPath, err)
	}

	var config DefaultConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("config.NewDefaultConfig: error parsing config file '%s': %w", configPath, err)
	}

	return &config, nil
}
