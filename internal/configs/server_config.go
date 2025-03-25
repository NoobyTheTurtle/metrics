package configs

type ServerConfig struct {
	ServerAddress string
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		ServerAddress: "localhost:8080",
	}
}
