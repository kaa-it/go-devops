package server

import (
	"flag"
	"os"
)

const (
	_serverAddress = ":8080"
)

type SelfConfig struct {
	Address string
}

type Config struct {
	Server SelfConfig
}

func NewConfig() *Config {
	address := flag.String(
		"a",
		_serverAddress,
		"server address as \"host:port\"",
	)

	flag.Parse()

	return &Config{
		Server: SelfConfig{
			Address: getEnv("ADDRESS", *address),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
