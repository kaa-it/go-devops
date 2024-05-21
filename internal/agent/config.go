package agent

import (
	"flag"
	"os"
	"strconv"
	"time"
)

const (
	_pollIntervalInSecs   = 2
	_reportIntervalInSecs = 10
	_serverAddress        = "localhost:8080"
)

type ServerConfig struct {
	Address string
}

type SelfConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	Key            string
}

type Config struct {
	Server ServerConfig
	Agent  SelfConfig
}

func NewConfig() *Config {
	address := flag.String(
		"a",
		_serverAddress,
		"server address",
	)
	reportInterval := flag.Int(
		"r",
		_reportIntervalInSecs,
		"report interval (seconds)",
	)
	pollInterval := flag.Int(
		"p",
		_pollIntervalInSecs,
		"poll interval (seconds)",
	)
	key := flag.String(
		"k",
		"",
		"hash key",
	)

	flag.Parse()

	pollDuration := time.Duration(getEnvInt("POLL_INTERVAL", *pollInterval)) * time.Second
	reportDuration := time.Duration(getEnvInt("REPORT_INTERVAL", *reportInterval)) * time.Second

	return &Config{
		Server: ServerConfig{
			Address: getEnv("ADDRESS", *address),
		},
		Agent: SelfConfig{
			PollInterval:   pollDuration,
			ReportInterval: reportDuration,
			Key:            getEnv("KEY", *key),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		val, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return defaultValue
		}

		return int(val)
	}

	return defaultValue
}
