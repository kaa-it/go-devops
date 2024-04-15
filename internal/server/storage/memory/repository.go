package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type gauges = map[string]float64
type counters = map[string]int64

var (
	ErrGaugeNotFound   = errors.New("gauge not found")
	ErrCounterNotFound = errors.New("counter not found")
	ErrNoConfig        = errors.New("no configuration found")
	ErrInvalidConfig   = errors.New("invalid configuration")
)

type StorageConfig struct {
	StoreInterval time.Duration
	StoreFilePath string
	Restore       bool
}

type fileStorage struct {
	Gauges   gauges   `json:"gauges"`
	Counters counters `json:"counters"`
}

type Storage struct {
	mu       sync.RWMutex
	gauges   gauges
	counters counters
	config   *StorageConfig
	wg       sync.WaitGroup
	done     chan struct{}
}

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

func (s *Storage) Wait() {
	close(s.done)

	s.wg.Wait()
}

func (s *Storage) UpdateGauge(name string, value float64) error {
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

func (s *Storage) UpdateCounter(name string, value int64) error {
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

func (s *Storage) ForEachGauge(fn func(key string, value float64)) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for key, value := range s.gauges {
		fn(key, value)
	}
}

func (s *Storage) ForEachCounter(fn func(key string, value int64)) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for key, value := range s.counters {
		fn(key, value)
	}
}

func (s *Storage) Gauge(name string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.gauges[name]
	if !ok {
		return 0, ErrGaugeNotFound
	}

	return value, nil
}

func (s *Storage) Counter(name string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.counters[name]
	if !ok {
		return 0, ErrCounterNotFound
	}

	return value, nil
}

func (s *Storage) TotalGauges() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.gauges)
}

func (s *Storage) TotalCounters() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.counters)
}

func (s *Storage) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.save()
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

// TODO: Запускаем горутину для сохраниения, читаем при старте; останов горутины; запись при завершении приложения;
// TODO: корректное сохранение горутины; запись всего синхронная, если период 0
