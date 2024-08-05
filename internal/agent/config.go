package agent

import (
	"encoding/json"
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
	Address string `json:"address"`
}

// SelfConfig contains configuration for metric client itself.
type SelfConfig struct {
	// PollInterval - interval for polling metrics
	PollInterval time.Duration `json:"poll_interval"`
	// ReportInterval - interval for sending reports to server.
	ReportInterval time.Duration `json:"report_interval"`
	// Key - cryptographic hash to encoding reports.
	Key string `json:"key"`
	// PublicKeyPath - path to file with public RSA key to encrypt requests
	PublicKeyPath *string `json:"crypto_key"`
}

// Config describes total configuration for metric agent.
type Config struct {
	// Server - configuration for metric server.
	Server ServerConfig `json:"server"`
	// Agent - configuration for agent itself.
	Agent SelfConfig `json:"agent"`
}

// NewConfig creates total configuration for metric agent.
func NewConfig() (*Config, error) {
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

	configPath := flag.String(
		"c",
		"",
		"path to file with agent confguration",
	)

	flag.Parse()

	pollDuration := time.Duration(getEnvInt("POLL_INTERVAL", *pollInterval)) * time.Second
	reportDuration := time.Duration(getEnvInt("REPORT_INTERVAL", *reportInterval)) * time.Second

	pubKeyPath := getEnv("CRYPTO_KEY", *publicKeyPath)

	var keyPath *string

	if pubKeyPath != "" {
		keyPath = &pubKeyPath
	}

	if *configPath == "" {
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
		}, nil
	}

	config, err := readConfig(*configPath)
	if err != nil {
		return nil, err
	}

	// TODO: Check env and flags and update config

	return config, nil
}

func readConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dec := json.NewDecoder(file)

	var config Config

	if err := dec.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
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
