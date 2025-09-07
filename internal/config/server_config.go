package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	ConfigPath    string
	ServerAddress string `env:"ADDRESS"`
	LogLevel      string `env:"LOG_LEVEL"`
	AppEnv        string `env:"APP_ENV"`
	Key           string `env:"KEY"`
	CryptoKey     string `env:"CRYPTO_KEY"`

	StoreInterval   uint   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`

	DatabaseDSN string `env:"DATABASE_DSN"`

	TrustedSubnet string `env:"TRUSTED_SUBNET"`

	GRPCServerAddress string `env:"GRPC_ADDRESS"`
	EnableGRPC        bool   `env:"ENABLE_GRPC"`
}

func NewServerConfig() (*ServerConfig, error) {
	config := &ServerConfig{
		ConfigPath: "configs/server.json",
	}

	if err := config.parseFlags(); err != nil {
		return nil, err
	}

	defaultConfig, err := NewServerDefaultConfig(config.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("config.NewServerConfig: loading default config from '%s': %w", config.ConfigPath, err)
	}

	if config.ServerAddress == "" {
		config.ServerAddress = defaultConfig.ServerAddress
	}
	if config.LogLevel == "" {
		config.LogLevel = defaultConfig.LogLevel
	}
	if config.AppEnv == "" {
		config.AppEnv = defaultConfig.AppEnv
	}
	if config.Key == "" {
		config.Key = defaultConfig.Key
	}
	if config.CryptoKey == "" {
		config.CryptoKey = defaultConfig.CryptoKey
	}
	if config.StoreInterval == 0 {
		config.StoreInterval = defaultConfig.StoreInterval
	}
	if config.FileStoragePath == "" {
		config.FileStoragePath = defaultConfig.FileStoragePath
	}
	if !config.Restore {
		config.Restore = defaultConfig.Restore
	}
	if config.DatabaseDSN == "" {
		config.DatabaseDSN = defaultConfig.DatabaseDSN
	}
	if config.TrustedSubnet == "" {
		config.TrustedSubnet = defaultConfig.TrustedSubnet
	}
	if config.GRPCServerAddress == "" {
		config.GRPCServerAddress = defaultConfig.GRPCServerAddress
	}
	if !config.EnableGRPC {
		config.EnableGRPC = defaultConfig.EnableGRPC
	}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("config.NewServerConfig: parsing environment variables: %w", err)
	}

	return config, nil
}

func (c *ServerConfig) parseFlags() error {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	fs.StringVar(&c.ConfigPath, "c", c.ConfigPath, "Path to config file")
	fs.StringVar(&c.ConfigPath, "config", c.ConfigPath, "Path to config file")
	fs.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Server address")
	fs.UintVar(&c.StoreInterval, "i", c.StoreInterval, "Store interval in seconds")
	fs.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "File storage path")
	fs.BoolVar(&c.Restore, "r", c.Restore, "Restore metrics from file storage")
	fs.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "PostgreSQL DSN")
	fs.StringVar(&c.Key, "k", c.Key, "Secret key for hashing")
	fs.StringVar(&c.CryptoKey, "crypto-key", c.CryptoKey, "Path to private key file for decryption")
	fs.StringVar(&c.TrustedSubnet, "t", c.TrustedSubnet, "Trusted subnet in CIDR notation")
	fs.StringVar(&c.GRPCServerAddress, "grpc-address", c.GRPCServerAddress, "gRPC server address")
	fs.BoolVar(&c.EnableGRPC, "enable-grpc", c.EnableGRPC, "Enable gRPC server")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("config.ServerConfig.parseFlags: %w", err)
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("config.ServerConfig.parseFlags: unknown command line arguments: %v", fs.Args())
	}

	return nil
}
