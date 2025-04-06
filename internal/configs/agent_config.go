package configs

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	ServerAddress  string        `env:"ADDRESS"`
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
	var pollIntervalSec, reportIntervalSec int

	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Server address")
	flag.IntVar(&pollIntervalSec, "p", int(c.PollInterval.Seconds()), "Poll interval in seconds")
	flag.IntVar(&reportIntervalSec, "r", int(c.ReportInterval.Seconds()), "Report interval in seconds")

	flag.Parse()

	c.PollInterval = time.Duration(pollIntervalSec) * time.Second
	c.ReportInterval = time.Duration(reportIntervalSec) * time.Second

	if flag.NArg() > 0 {
		return fmt.Errorf("unknown command line arguments: %v", flag.Args())
	}

	return nil
}
