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

	StoreInterval   uint   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
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

		StoreInterval:   defaultConfig.StoreInterval,
		FileStoragePath: defaultConfig.FileStoragePath,
		Restore:         defaultConfig.Restore,
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
	flag.UintVar(&c.StoreInterval, "i", c.StoreInterval, "Store interval in seconds")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "File storage path")
	flag.BoolVar(&c.Restore, "r", c.Restore, "Restore metrics from file storage")

	flag.Parse()

	if flag.NArg() > 0 {
		return fmt.Errorf("unknown command line arguments: %v", flag.Args())
	}

	return nil
}
