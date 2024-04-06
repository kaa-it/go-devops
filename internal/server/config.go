package server

import (
	"flag"
	"os"
)

const (
	_serverAddress = ":8080"
	_logLevel      = "info"
)

type SelfConfig struct {
	Address  string
	LogLevel string
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

	logLevel := flag.String(
		"l",
		_logLevel,
		"log level",
	)

	flag.Parse()

	return &Config{
		Server: SelfConfig{
			Address:  getEnv("ADDRESS", *address),
			LogLevel: getEnv("LOG_LEVEL", *logLevel),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
