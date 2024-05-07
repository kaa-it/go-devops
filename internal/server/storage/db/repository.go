package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrGaugeNotFound   = errors.New("gauge not found")
	ErrCounterNotFound = errors.New("counter not found")
	ErrNoConfig        = errors.New("no configuration found")
	ErrInvalidConfig   = errors.New("invalid configuration")
)

type StorageConfig struct {
	DSN string
}

type Storage struct {
	config *StorageConfig
	dbpool *pgxpool.Pool
}

func NewStorage(config *StorageConfig) (*Storage, error) {
	if config == nil {
		return nil, ErrNoConfig
	}

	dbpool, err := pgxpool.New(context.Background(), config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &Storage{
		config,
		dbpool,
	}, nil
}

func (s *Storage) Close() {
	s.dbpool.Close()
}

func (s *Storage) Ping() error {
	var greeting string

	return s.dbpool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
}
