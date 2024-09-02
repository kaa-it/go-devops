package agent

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"strconv"
	"time"
)

const (
	_pollIntervalInSecs   = 2
	_reportIntervalInSecs = 10
)

type configFile struct {
	PollInterval   int    `json:"poll_interval"`
	ReportInterval int    `json:"report_interval"`
	Address        string `json:"address"`
	Key            string `json:"key"`
	PublicKeyPath  string `json:"crypto_key"`
	GRPCAddress    string `json:"grpc_address"`
}

// ServerConfig contains configuration if metric server
type ServerConfig struct {
	// Address - address of metric server.
	Address string
	// GRPSAddress - address of GRPC metric server.
	GRPCAddress string
}

// SelfConfig contains configuration for metric client itself.
type SelfConfig struct {
	// PollInterval - interval for polling metrics.
	PollInterval time.Duration
	// ReportInterval - interval for sending reports to server.
	ReportInterval time.Duration
	// Key - cryptographic hash to encoding reports.
	Key string
	// PublicKeyPath - path to file with public RSA key to encrypt requests.
	PublicKeyPath string
}

// Config describes total configuration for metric agent.
type Config struct {
	// Server - configuration for metric server.
	Server ServerConfig
	// Agent - configuration for agent itself.
	Agent SelfConfig
}

// NewConfig creates total configuration for metric agent.
func NewConfig() (*Config, error) {
	address := flag.String(
		"a",
		"",
		"server address",
	)

	reportInterval := flag.Int(
		"r",
		-1,
		"report interval (seconds)",
	)

	pollInterval := flag.Int(
		"p",
		-1,
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
		"path to file with agent configuration",
	)

	grpcAddress := flag.String(
		"g",
		"",
		"address of GRPC server",
	)

	flag.Parse()

	configFilePath := getEnv("CONFIG", *configPath)

	config := configFile{
		PollInterval:   _pollIntervalInSecs,
		ReportInterval: _reportIntervalInSecs,
		Address:        "",
		Key:            "",
		PublicKeyPath:  "",
		GRPCAddress:    "",
	}

	if configFilePath != "" {
		if err := readConfig(configFilePath, &config); err != nil {
			return nil, err
		}
	}

	if *address != "" {
		config.Address = *address
	}

	if *reportInterval != -1 {
		config.ReportInterval = *reportInterval
	}

	if *pollInterval != -1 {
		config.PollInterval = *pollInterval
	}

	if *key != "" {
		config.Key = *key
	}

	if *publicKeyPath != "" {
		config.PublicKeyPath = *publicKeyPath
	}

	if *grpcAddress != "" {
		config.GRPCAddress = *grpcAddress
	}

	pollDuration := time.Duration(getEnvInt("POLL_INTERVAL", config.PollInterval)) * time.Second
	reportDuration := time.Duration(getEnvInt("REPORT_INTERVAL", config.ReportInterval)) * time.Second

	configuration := &Config{
		Server: ServerConfig{
			Address:     getEnv("ADDRESS", config.Address),
			GRPCAddress: getEnv("GRPC_ADDRESS", config.GRPCAddress),
		},
		Agent: SelfConfig{
			PollInterval:   pollDuration,
			ReportInterval: reportDuration,
			Key:            getEnv("KEY", config.Key),
			PublicKeyPath:  getEnv("CRYPTO_KEY", config.PublicKeyPath),
		},
	}

	if configuration.Server.Address == "" && configuration.Server.GRPCAddress == "" {
		return nil, errors.New("no server address specified")
	}

	if configuration.Server.Address != "" && configuration.Server.GRPCAddress != "" {
		return nil, errors.New("cannot specify both server address and gRPC address")
	}

	return configuration, nil
}

func readConfig(configPath string, config *configFile) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	dec := json.NewDecoder(file)

	if err := dec.Decode(config); err != nil {
		return err
	}

	return nil
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
