package configs

import (
	"flag"
	"fmt"
)

type ServerConfig struct {
	ServerAddress string
}

func NewServerConfig() (*ServerConfig, error) {
	config := &ServerConfig{
		ServerAddress: "localhost:8080",
	}

	if err := config.parseFlags(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *ServerConfig) parseFlags() error {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Server address (default: localhost:8080)")

	flag.Parse()

	if flag.NArg() > 0 {
		return fmt.Errorf("unknown command line arguments: %v", flag.Args())
	}

	return nil
}
