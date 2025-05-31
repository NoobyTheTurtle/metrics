package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	PollInterval   uint   `env:"POLL_INTERVAL"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	ServerAddress  string `env:"ADDRESS"`
	LogLevel       string `env:"LOG_LEVEL"`
	AppEnv         string `env:"APP_ENV"`
	Key            string `env:"KEY"`
	RateLimit      uint   `env:"RATE_LIMIT"`
}

func NewAgentConfig(configPath string) (*AgentConfig, error) {
	defaultConfig, err := NewDefaultConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("config.NewAgentConfig: loading default config from '%s': %w", configPath, err)
	}

	config := &AgentConfig{
		PollInterval:   defaultConfig.PollInterval,
		ReportInterval: defaultConfig.ReportInterval,
		ServerAddress:  defaultConfig.ServerAddress,
		LogLevel:       defaultConfig.LogLevel,
		AppEnv:         defaultConfig.AppEnv,
		Key:            defaultConfig.Key,
		RateLimit:      defaultConfig.RateLimit,
	}

	if err := config.parseFlags(); err != nil {
		return nil, err
	}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("config.NewAgentConfig: parsing environment variables: %w", err)
	}

	return config, nil
}

func (c *AgentConfig) parseFlags() error {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Server address")
	flag.UintVar(&c.PollInterval, "p", c.PollInterval, "Poll interval in seconds")
	flag.UintVar(&c.ReportInterval, "r", c.ReportInterval, "Report interval in seconds")
	flag.StringVar(&c.Key, "k", c.Key, "Secret key for hashing")
	flag.UintVar(&c.RateLimit, "l", c.RateLimit, "Rate limit for concurrent requests")

	flag.Parse()

	if flag.NArg() > 0 {
		return fmt.Errorf("config.AgentConfig.parseFlags: unknown command line arguments: %v", flag.Args())
	}

	return nil
}
