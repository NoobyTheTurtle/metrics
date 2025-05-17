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
	Key           string `env:"KEY"`

	StoreInterval   uint   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`

	DatabaseDSN string `env:"DATABASE_DSN"`
}

func NewServerConfig(configPath string) (*ServerConfig, error) {
	defaultConfig, err := NewDefaultConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("config.NewServerConfig: loading default config from '%s': %w", configPath, err)
	}

	config := &ServerConfig{
		ServerAddress: defaultConfig.ServerAddress,
		LogLevel:      defaultConfig.LogLevel,
		AppEnv:        defaultConfig.AppEnv,
		Key:           defaultConfig.Key,

		StoreInterval:   defaultConfig.StoreInterval,
		FileStoragePath: defaultConfig.FileStoragePath,
		Restore:         defaultConfig.Restore,

		DatabaseDSN: defaultConfig.DatabaseDSN,
	}

	if err := config.parseFlags(); err != nil {
		return nil, err
	}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("config.NewServerConfig: parsing environment variables: %w", err)
	}

	return config, nil
}

func (c *ServerConfig) parseFlags() error {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Server address")
	flag.UintVar(&c.StoreInterval, "i", c.StoreInterval, "Store interval in seconds")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "File storage path")
	flag.BoolVar(&c.Restore, "r", c.Restore, "Restore metrics from file storage")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "PostgreSQL DSN")
	flag.StringVar(&c.Key, "k", c.Key, "Secret key for hashing")

	flag.Parse()

	if flag.NArg() > 0 {
		return fmt.Errorf("config.ServerConfig.parseFlags: unknown command line arguments: %v", flag.Args())
	}

	return nil
}
