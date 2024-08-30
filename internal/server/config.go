package server

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/kaa-it/go-devops/internal/server/storage/db"
	"github.com/kaa-it/go-devops/internal/server/storage/memory"
)

const (
	_serverAddress       = ":8080"
	_logLevel            = "info"
	_storeIntervalInSecs = 300
	_storeFilePath       = "/tmp/metrics-db.json"
	_restore             = true
)

type configFile struct {
	Address        string `json:"address"`
	Restore        bool   `json:"restore"`
	StoreInterval  int    `json:"store_interval"`
	StoreFilePath  string `json:"store_file"`
	DatabaseDSN    string `json:"database_dsn"`
	Key            string `json:"key"`
	PrivateKeyPath string `json:"crypto_key"`
	LogLevel       string `json:"log_level"`
	TrustedSubnet  string `json:"trusted_subnet"`
}

// SelfConfig contains configuration for the server itself.
type SelfConfig struct {
	// Address - address to listen by server.
	Address string
	// LogLevel - minimal level for log of server.
	LogLevel string
	// Key - cryptographic key for decoding update requests.
	Key string
	// PrivateKeyPath - path to file with private RSA key to dencrypt requests
	PrivateKeyPath string
	// TrustedSubnet - trusted subnet CIDR
	TrustedSubnet string
}

// Config contains total configuration for server.
type Config struct {
	// Server - configuration for server itself.
	Server SelfConfig
	// Storage - configuration for memory storage.
	Storage memory.StorageConfig
	// DBStorage - configuration for database storage.
	DBStorage db.StorageConfig
}

// NewConfig creates total server configuration.
func NewConfig() (*Config, error) {
	address := flag.String(
		"a",
		"",
		"server address as \"host:port\"",
	)

	logLevel := flag.String(
		"l",
		"",
		"log level",
	)

	storeInterval := flag.Int(
		"i",
		-1,
		"store interval (seconds)",
	)

	storeFilePath := flag.String(
		"f",
		"",
		"store file path",
	)

	restore := flag.Bool(
		"r",
		_restore,
		"restore metrics",
	)

	dsn := flag.String(
		"d",
		"",
		"database DSN",
	)

	key := flag.String(
		"k",
		"",
		"hash key",
	)

	privateKeyPath := flag.String(
		"crypto-key",
		"",
		"path to file with RSA private crypto key",
	)

	configPath := flag.String(
		"c",
		"",
		"path to file with server configuration",
	)

	trustedSubnet := flag.String(
		"t",
		"",
		"trusted subnet CIDR",
	)

	flag.Parse()

	configFilePath := getEnv("CONFIG", *configPath)

	config := configFile{
		Address:        _serverAddress,
		Restore:        _restore,
		StoreInterval:  _storeIntervalInSecs,
		StoreFilePath:  _storeFilePath,
		DatabaseDSN:    "",
		Key:            "",
		PrivateKeyPath: "",
		LogLevel:       _logLevel,
		TrustedSubnet:  "",
	}

	if configFilePath != "" {
		if err := readConfig(configFilePath, &config); err != nil {
			return nil, err
		}
	}

	if *address != "" {
		config.Address = *address
	}

	if *restore != config.Restore {
		config.Restore = *restore
	}

	if *storeInterval != -1 {
		config.StoreInterval = *storeInterval
	}

	if *storeFilePath != "" {
		config.StoreFilePath = *storeFilePath
	}

	if *dsn != "" {
		config.DatabaseDSN = *dsn
	}

	if *key != "" {
		config.Key = *key
	}

	if *privateKeyPath != "" {
		config.PrivateKeyPath = *privateKeyPath
	}

	if *logLevel != "" {
		config.LogLevel = *logLevel
	}

	if *trustedSubnet != "" {
		config.TrustedSubnet = *trustedSubnet
	}

	storeDuration := time.Duration(getEnvInt("STORE_INTERVAL", config.StoreInterval)) * time.Second

	return &Config{
		Server: SelfConfig{
			Address:        getEnv("ADDRESS", config.Address),
			LogLevel:       getEnv("LOG_LEVEL", config.LogLevel),
			Key:            getEnv("KEY", config.Key),
			PrivateKeyPath: getEnv("CRYPTO_KEY", config.PrivateKeyPath),
			TrustedSubnet:  getEnv("TRUSTED_SUBNET", config.TrustedSubnet),
		},
		Storage: memory.StorageConfig{
			StoreInterval: storeDuration,
			StoreFilePath: getEnv("FILE_STORAGE_PATH", config.StoreFilePath),
			Restore:       getEnvBool("RESTORE", config.Restore),
		},
		DBStorage: db.StorageConfig{
			DSN: getEnv("DATABASE_DSN", config.DatabaseDSN),
		},
	}, nil
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

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		val, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}

		return val
	}

	return defaultValue
}
