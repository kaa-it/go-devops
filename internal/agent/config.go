package agent

import "os"

type ServerConfig struct {
	Host string
	Port string
}

type Config struct {
	Server ServerConfig
}

func NewConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "127.0.0.1"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
