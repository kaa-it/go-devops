package memory

import (
	"errors"
	"sync"
)

type gauges = map[string]float64
type counters = map[string]int64

var (
	ErrGaugeNotFound   = errors.New("gauge not found")
	ErrCounterNotFound = errors.New("counter not found")
)

type Storage struct {
	mu       sync.RWMutex
	gauges   gauges
	counters counters
}

func NewStorage() *Storage {
	return &Storage{
		gauges:   make(gauges),
		counters: make(counters),
	}
}

func (s *Storage) UpdateGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
}

func (s *Storage) UpdateCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += value
}

func (s *Storage) ForEachGauge(fn func(key string, value float64)) {
	s.mu.RLock()

	for key, value := range s.gauges {
		s.mu.RUnlock()

		fn(key, value)

		s.mu.RLock()
	}

	s.mu.RUnlock()
}

func (s *Storage) ForEachCounter(fn func(key string, value int64)) {
	s.mu.RLock()

	for key, value := range s.counters {
		s.mu.RUnlock()

		fn(key, value)

		s.mu.RLock()
	}

	s.mu.RUnlock()
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
