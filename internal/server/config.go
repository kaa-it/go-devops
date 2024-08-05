package server

import (
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

// SelfConfig contains configuration for the server itself.
type SelfConfig struct {
	// Address - address to listen by server.
	Address string
	// LogLevel - minimal level for log of server.
	LogLevel string
	// Key - cryptographic key for decoding update requests.
	Key string
	// PrivateKeyPath - path to file with private RSA key to dencrypt requests
	PrivateKeyPath *string
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

	storeInterval := flag.Int(
		"i",
		_storeIntervalInSecs,
		"store interval (seconds)",
	)

	storeFilePath := flag.String(
		"f",
		_storeFilePath,
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

	flag.Parse()

	storeDuration := time.Duration(getEnvInt("STORE_INTERVAL", *storeInterval)) * time.Second

	privKeyPath := getEnv("CRYPTO_KEY", *privateKeyPath)

	var keyPath *string

	if privKeyPath != "" {
		keyPath = &privKeyPath
	}

	return &Config{
		Server: SelfConfig{
			Address:        getEnv("ADDRESS", *address),
			LogLevel:       getEnv("LOG_LEVEL", *logLevel),
			Key:            getEnv("KEY", *key),
			PrivateKeyPath: keyPath,
		},
		Storage: memory.StorageConfig{
			StoreInterval: storeDuration,
			StoreFilePath: getEnv("FILE_STORAGE_PATH", *storeFilePath),
			Restore:       getEnvBool("RESTORE", *restore),
		},
		DBStorage: db.StorageConfig{
			DSN: getEnv("DATABASE_DSN", *dsn),
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
