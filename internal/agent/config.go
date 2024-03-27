package agent

import "os"

type ServerConfig struct {
	Address string
}

type Config struct {
	Server ServerConfig
}

func NewConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Address: getEnv("ADDRESS", "127.0.0.1:8080"),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
