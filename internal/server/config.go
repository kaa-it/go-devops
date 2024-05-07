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

type SelfConfig struct {
	Address  string
	LogLevel string
}

type Config struct {
	Server    SelfConfig
	Storage   memory.StorageConfig
	DbStorage db.StorageConfig
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

	flag.Parse()

	storeDuration := time.Duration(getEnvInt("STORE_INTERVAL", *storeInterval)) * time.Second

	return &Config{
		Server: SelfConfig{
			Address:  getEnv("ADDRESS", *address),
			LogLevel: getEnv("LOG_LEVEL", *logLevel),
		},
		Storage: memory.StorageConfig{
			StoreInterval: storeDuration,
			StoreFilePath: getEnv("FILE_STORAGE_PATH", *storeFilePath),
			Restore:       getEnvBool("RESTORE", *restore),
		},
		DbStorage: db.StorageConfig{
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
