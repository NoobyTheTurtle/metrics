package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DefaultConfig struct {
	PollInterval   uint   `yaml:"poll_interval"`
	ReportInterval uint   `yaml:"report_interval"`
	ServerAddress  string `yaml:"server_address"`
	LogLevel       string `yaml:"log_level"`
	AppEnv         string `yaml:"app_env"`
}

func NewDefaultConfig(configPath string) (*DefaultConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config DefaultConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}
