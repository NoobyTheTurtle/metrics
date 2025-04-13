package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	ServerAddress string `env:"ADDRESS"`
	LogLevel      string `env:"LOG_LEVEL"`
	AppEnv        string `env:"APP_ENV"`
}

func NewServerConfig(configPath string) (*ServerConfig, error) {
	defaultConfig, err := NewDefaultConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading default config: %w", err)
	}

	config := &ServerConfig{
		ServerAddress: defaultConfig.ServerAddress,
		LogLevel:      defaultConfig.LogLevel,
		AppEnv:        defaultConfig.AppEnv,
	}

	if err := config.parseFlags(); err != nil {
		return nil, err
	}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("parsing environment variables: %w", err)
	}

	return config, nil
}

func (c *ServerConfig) parseFlags() error {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Server address")

	flag.Parse()

	if flag.NArg() > 0 {
		return fmt.Errorf("unknown command line arguments: %v", flag.Args())
	}

	return nil
}
