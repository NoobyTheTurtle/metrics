package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	ConfigPath        string
	PollInterval      uint   `env:"POLL_INTERVAL"`
	ReportInterval    uint   `env:"REPORT_INTERVAL"`
	ServerAddress     string `env:"ADDRESS"`
	LogLevel          string `env:"LOG_LEVEL"`
	AppEnv            string `env:"APP_ENV"`
	Key               string `env:"KEY"`
	RateLimit         uint   `env:"RATE_LIMIT"`
	CryptoKey         string `env:"CRYPTO_KEY"`
	GRPCServerAddress string `env:"GRPC_ADDRESS"`
	EnableGRPC        bool   `env:"ENABLE_GRPC"`
}

func NewAgentConfig() (*AgentConfig, error) {
	config := &AgentConfig{
		ConfigPath: "configs/agent.json",
	}

	if err := config.parseFlags(); err != nil {
		return nil, err
	}

	defaultConfig, err := NewAgentDefaultConfig(config.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("config.NewAgentConfig: loading default config from '%s': %w", config.ConfigPath, err)
	}

	if config.PollInterval == 0 {
		config.PollInterval = defaultConfig.PollInterval
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = defaultConfig.ReportInterval
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
	if config.RateLimit == 0 {
		config.RateLimit = defaultConfig.RateLimit
	}
	if config.CryptoKey == "" {
		config.CryptoKey = defaultConfig.CryptoKey
	}
	if config.GRPCServerAddress == "" {
		config.GRPCServerAddress = defaultConfig.GRPCServerAddress
	}
	if !config.EnableGRPC {
		config.EnableGRPC = defaultConfig.EnableGRPC
	}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("config.NewAgentConfig: parsing environment variables: %w", err)
	}

	return config, nil
}

func (c *AgentConfig) parseFlags() error {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	fs.StringVar(&c.ConfigPath, "c", c.ConfigPath, "Path to config file")
	fs.StringVar(&c.ConfigPath, "config", c.ConfigPath, "Path to config file")
	fs.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Server address")
	fs.UintVar(&c.PollInterval, "p", c.PollInterval, "Poll interval in seconds")
	fs.UintVar(&c.ReportInterval, "r", c.ReportInterval, "Report interval in seconds")
	fs.StringVar(&c.Key, "k", c.Key, "Secret key for hashing")
	fs.UintVar(&c.RateLimit, "l", c.RateLimit, "Rate limit for concurrent requests")
	fs.StringVar(&c.CryptoKey, "crypto-key", c.CryptoKey, "Path to public key file for encryption")
	fs.StringVar(&c.GRPCServerAddress, "grpc-address", c.GRPCServerAddress, "gRPC server address")
	fs.BoolVar(&c.EnableGRPC, "enable-grpc", c.EnableGRPC, "Enable gRPC transport")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("config.AgentConfig.parseFlags: %w", err)
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("config.AgentConfig.parseFlags: unknown command line arguments: %v", fs.Args())
	}

	return nil
}
