// Package memory contains in-memory implementation of metric storage.
package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/kaa-it/go-devops/internal/api"
)

type gauges = map[string]float64
type counters = map[string]int64

// Sentinel errors for in-memory storage.
var (
	ErrGaugeNotFound   = errors.New("gauge not found")
	ErrCounterNotFound = errors.New("counter not found")
	ErrNoConfig        = errors.New("no configuration found")
	ErrInvalidConfig   = errors.New("invalid configuration")
)

// StorageConfig describes configuration of in-memory storage.
type StorageConfig struct {
	// StoreInterval - interval for backup storage in file.
	StoreInterval time.Duration
	// StoreFilePath - path to file for storage backup.
	StoreFilePath string
	// Restore - if true storage will be restored from backup at start of application.
	Restore bool
}

type fileStorage struct {
	Gauges   gauges   `json:"gauges"`
	Counters counters `json:"counters"`
}

// Storage describes in-memory storage.
type Storage struct {
	mu       sync.RWMutex
	gauges   gauges
	counters counters
	config   *StorageConfig
	wg       sync.WaitGroup
	done     chan struct{}
}

// NewStorage creates new in-memory storage instance with given configuration.
//
// If config is nil returns ErrNoConfig.
// If backup enabled but config.StoreFilePath is empty returns ErrInvalidConfig.
func NewStorage(config *StorageConfig) (*Storage, error) {
	if config == nil {
		return nil, ErrNoConfig
	}

	if (config.Restore || config.StoreInterval != 0) && config.StoreFilePath == "" {
		return nil, ErrInvalidConfig
	}

	s := &Storage{
		config: config,
		done:   make(chan struct{}),
	}

	if config.Restore {
		data, err := load(config.StoreFilePath)
		if err != nil {
			return nil, fmt.Errorf("restore failed: %w", err)
		}

		s.gauges = data.Gauges
		s.counters = data.Counters
	} else {
		s.gauges = make(gauges)
		s.counters = make(counters)
	}

	if config.StoreInterval != 0 {
		s.wg.Add(1)
		go s.saver()
	}

	return s, nil
}

func (s *Storage) saver() {
	defer s.wg.Done()

	for {
		select {
		case <-time.After(s.config.StoreInterval):
			s.Save()
		case <-s.done:
			return
		}
	}
}

// Wait waits for completion of backup goroutine.
func (s *Storage) Wait() {
	close(s.done)

	s.wg.Wait()
}

// UpdateGauge updates gauge metric with given name in hashmap.
//
// May return output errors if backup enabled. Thread-safe.
func (s *Storage) UpdateGauge(_ context.Context, name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value

	if s.config.StoreInterval == 0 {
		if err := s.save(); err != nil {
			return err
		}
	}

	return nil
}

// UpdateCounter updates counter metric with given name in hashmap.
//
// May return output errors if backup enabled. Thread-safe.
func (s *Storage) UpdateCounter(_ context.Context, name string, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += value

	if s.config.StoreInterval == 0 {
		if err := s.save(); err != nil {
			return err
		}
	}

	return nil

}

// ForEachGauge applies given function to every gauge metric in storage. Thread-safe.
func (s *Storage) ForEachGauge(_ context.Context, fn func(key string, value float64)) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for key, value := range s.gauges {
		fn(key, value)
	}

	return nil
}

// ForEachCounter applies given function to every counter metric in storage. Thread-safe.
func (s *Storage) ForEachCounter(_ context.Context, fn func(key string, value int64)) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for key, value := range s.counters {
		fn(key, value)
	}

	return nil
}

// Gauge returns value of gauge metric by its name.
//
// If metric with given name is not found returns ErrGaugeNotFound.
//
// Method is thread-safe.
func (s *Storage) Gauge(_ context.Context, name string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.gauges[name]
	if !ok {
		return 0, ErrGaugeNotFound
	}

	return value, nil
}

// Counter returns value of counter metric by its name.
//
// If metric with given name is not found returns ErrCounterNotFound.
//
// Method is thread-safe.
func (s *Storage) Counter(_ context.Context, name string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.counters[name]
	if !ok {
		return 0, ErrCounterNotFound
	}

	return value, nil
}

// TotalGauges returns total amount of gauge metric from storage. Thread-safe.
func (s *Storage) TotalGauges(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.gauges), nil
}

// TotalCounters returns total amount of counter metric from storage. Thread-safe.
func (s *Storage) TotalCounters(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.counters), nil
}

// Save creates backup of storage.
func (s *Storage) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.save()
}

// Updates updates some metrics simultaneously in storage. Thread-safe.
func (s *Storage) Updates(_ context.Context, metrics []api.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, m := range metrics {
		if m.MType == api.CounterType {
			s.counters[m.ID] += *m.Delta
		} else {
			s.gauges[m.ID] = *m.Value
		}
	}

	if s.config.StoreInterval == 0 {
		if err := s.save(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) save() error {
	if s.config.StoreFilePath == "" {
		return nil
	}

	file, err := os.OpenFile(s.config.StoreFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	data := fileStorage{
		Gauges:   s.gauges,
		Counters: s.counters,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

func load(storeFilePath string) (*fileStorage, error) {
	file, err := os.Open(storeFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &fileStorage{
				Gauges:   make(gauges),
				Counters: make(counters),
			}, nil
		}

		return nil, err
	}

	defer file.Close()

	var data fileStorage

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}
