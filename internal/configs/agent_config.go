package configs

import (
	"flag"
	"fmt"
	"time"
)

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddress  string
}

func NewAgentConfig() (*AgentConfig, error) {
	config := &AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerAddress:  "localhost:8080",
	}

	if err := config.parseFlags(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *AgentConfig) parseFlags() error {
	var pollIntervalSec, reportIntervalSec int

	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Server address (default: http://localhost:8080)")
	flag.IntVar(&pollIntervalSec, "p", int(c.PollInterval.Seconds()), "Poll interval in seconds (default: 2)")
	flag.IntVar(&reportIntervalSec, "r", int(c.ReportInterval.Seconds()), "Report interval in seconds (default: 10)")

	flag.Parse()

	c.PollInterval = time.Duration(pollIntervalSec) * time.Second
	c.ReportInterval = time.Duration(reportIntervalSec) * time.Second

	if flag.NArg() > 0 {
		return fmt.Errorf("unknown command line arguments: %v", flag.Args())
	}

	return nil
}
