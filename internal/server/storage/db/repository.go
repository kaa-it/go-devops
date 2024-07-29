// Package db contains Postgresql database implementation of metric storage.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kaa-it/go-devops/internal/api"
)

// Sentinel errors for database storage.
var (
	ErrGaugeNotFound   = errors.New("gauge not found")
	ErrCounterNotFound = errors.New("counter not found")
	ErrNoConfig        = errors.New("no configuration found")
)

// StorageConfig describes configuration of database storage.
type StorageConfig struct {
	DSN string
}

// Storage describes database storage.
type Storage struct {
	config *StorageConfig
	dbpool *pgxpool.Pool
}

// NewStorage creates new instance of database storage.
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

// Close closes database connections.
func (s *Storage) Close() {
	s.dbpool.Close()
}

// Ping tests database connection.
func (s *Storage) Ping(ctx context.Context) error {
	return s.dbpool.Ping(ctx)
}

// Initialize initializes tables in database. Method is idempotent.
func (s *Storage) Initialize(ctx context.Context) error {
	_, err := s.dbpool.Exec(
		ctx,
		"CREATE TABLE IF NOT EXISTS gauges"+
			" (name TEXT PRIMARY KEY, value DOUBLE PRECISION NOT NULL)",
	)

	if err != nil {
		return err
	}

	_, err = s.dbpool.Exec(
		ctx,
		"CREATE TABLE IF NOT EXISTS counters"+
			" (name TEXT PRIMARY KEY, value BIGINT NOT NULL)",
	)

	return err
}

// UpdateGauge updates gauge metric with given name in database.
func (s *Storage) UpdateGauge(ctx context.Context, name string, value float64) error {
	_, err := s.dbpool.Exec(
		ctx,
		"INSERT INTO gauges (name, value) VALUES (@name, @value)"+
			" ON CONFLICT (name) DO UPDATE"+
			" SET value = EXCLUDED.value",
		pgx.NamedArgs{
			"name":  name,
			"value": value,
		},
	)

	return err
}

// UpdateCounter updates counter metric with given name in database.
func (s *Storage) UpdateCounter(ctx context.Context, name string, value int64) error {
	_, err := s.dbpool.Exec(
		ctx,
		"INSERT INTO counters (name, value) VALUES (@name, @value)"+
			" ON CONFLICT (name) DO UPDATE"+
			" SET value = EXCLUDED.value + counters.value",
		pgx.NamedArgs{
			"name":  name,
			"value": value,
		},
	)

	return err
}

// Gauge returns value of gauge metric by its name.
//
// If metric with given name is not found returns ErrGaugeNotFound
func (s *Storage) Gauge(ctx context.Context, name string) (float64, error) {
	var value float64

	err := s.dbpool.QueryRow(
		ctx,
		"SELECT value FROM gauges WHERE name = @name",
		pgx.NamedArgs{
			"name": name,
		},
	).Scan(&value)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrGaugeNotFound
	}

	return value, err
}

// Counter returns value of counter metric by its name.
//
// If metric with given name is not found returns ErrCounterNotFound.
func (s *Storage) Counter(ctx context.Context, name string) (int64, error) {
	var value int64

	err := s.dbpool.QueryRow(
		ctx,
		"SELECT value FROM counters WHERE name = @name",
		pgx.NamedArgs{
			"name": name,
		},
	).Scan(&value)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrCounterNotFound
	}

	return value, err
}

// ForEachGauge applies given function to every gauge metric in database.
func (s *Storage) ForEachGauge(ctx context.Context, fn func(name string, value float64)) error {
	rows, err := s.dbpool.Query(
		ctx,
		"SELECT * FROM gauges",
	)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return err
		}

		fn(name, value)
	}

	return nil
}

// ForEachCounter applies given function to every counter metric in database.
func (s *Storage) ForEachCounter(ctx context.Context, fn func(name string, value int64)) error {
	rows, err := s.dbpool.Query(
		ctx,
		"SELECT * FROM counters",
	)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			return err
		}

		fn(name, value)
	}

	return nil
}

// TotalCounters returns total amount of counter metric from database.
func (s *Storage) TotalCounters(ctx context.Context) (int, error) {
	var value int

	err := s.dbpool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM counters",
	).Scan(&value)

	return value, err
}

// TotalGauges returns total amount of gauge metric from database.
func (s *Storage) TotalGauges(ctx context.Context) (int, error) {
	var value int

	err := s.dbpool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM gauges",
	).Scan(&value)

	return value, err
}

// Updates does batch update of metrics in database.
func (s *Storage) Updates(ctx context.Context, metrics []api.Metrics) error {
	queryGauge := "INSERT INTO gauges (name, value) VALUES (@name, @value)" +
		" ON CONFLICT (name) DO UPDATE" +
		" SET value = EXCLUDED.value"

	queryCounters := "INSERT INTO counters (name, value) VALUES (@name, @value)" +
		" ON CONFLICT (name) DO UPDATE" +
		" SET value = EXCLUDED.value + counters.value"

	batch := &pgx.Batch{}
	for _, metric := range metrics {
		if metric.MType == api.CounterType {
			args := pgx.NamedArgs{
				"name":  metric.ID,
				"value": metric.Delta,
			}
			batch.Queue(queryCounters, args)
		} else {
			args := pgx.NamedArgs{
				"name":  metric.ID,
				"value": metric.Value,
			}
			batch.Queue(queryGauge, args)
		}
	}

	results := s.dbpool.SendBatch(ctx, batch)
	defer results.Close()

	_, err := results.Exec()

	return err
}
