package configs

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	PollInterval   uint   `env:"POLL_INTERVAL"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	ServerAddress  string `env:"ADDRESS"`
}

func NewAgentConfig() (*AgentConfig, error) {
	config := &AgentConfig{
		PollInterval:   DefaultPollInterval,
		ReportInterval: DefaultReportInterval,
		ServerAddress:  DefaultServerAddress,
	}

	if err := config.parseFlags(); err != nil {
		return nil, err
	}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("parsing environment variables: %w", err)
	}

	return config, nil
}

func (c *AgentConfig) parseFlags() error {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Server address")
	flag.UintVar(&c.PollInterval, "p", c.PollInterval, "Poll interval in seconds")
	flag.UintVar(&c.ReportInterval, "r", c.ReportInterval, "Report interval in seconds")

	flag.Parse()

	if flag.NArg() > 0 {
		return fmt.Errorf("unknown command line arguments: %v", flag.Args())
	}

	return nil
}
