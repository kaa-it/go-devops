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

// ServerConfig contains configuration if metric server
type ServerConfig struct {
	// Address - address of metric server.
	Address string
}

// SelfConfig contains configuration for metric client itself.
type SelfConfig struct {
	// PollInterval - interval for polling metrics
	PollInterval time.Duration
	// ReportInterval - interval for sending reports to server.
	ReportInterval time.Duration
	// Key - cryptographic hash to encoding reports.
	Key string
	// PublicKeyPath - path to file with public RSA key to encrypt requests
	PublicKeyPath *string
}

// Config describes total configuration for metric agent.
type Config struct {
	// Server - configuration for metric server.
	Server ServerConfig
	// Agent - configuration for agent itself.
	Agent SelfConfig
}

// NewConfig creates total configuration for metric agent.
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

	publicKeyPath := flag.String(
		"crypto-key",
		"",
		"path to file with RSA public crypto key",
	)

	flag.Parse()

	pollDuration := time.Duration(getEnvInt("POLL_INTERVAL", *pollInterval)) * time.Second
	reportDuration := time.Duration(getEnvInt("REPORT_INTERVAL", *reportInterval)) * time.Second

	pubKeyPath := getEnv("CRYPTO_KEY", *publicKeyPath)

	var keyPath *string

	if pubKeyPath != "" {
		keyPath = &pubKeyPath
	}

	return &Config{
		Server: ServerConfig{
			Address: getEnv("ADDRESS", *address),
		},
		Agent: SelfConfig{
			PollInterval:   pollDuration,
			ReportInterval: reportDuration,
			Key:            getEnv("KEY", *key),
			PublicKeyPath:  keyPath,
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
