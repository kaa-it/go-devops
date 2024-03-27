package agent

import "sync"

type Gauges = map[string]float64
type Counters = map[string]int64

type Storage struct {
	mu       sync.Mutex
	gauges   Gauges
	counters Counters
}

func NewStorage() *Storage {
	return &Storage{
		gauges:   make(Gauges),
		counters: make(Counters),
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
	s.mu.Lock()

	for key, value := range s.gauges {
		s.mu.Unlock()

		fn(key, value)

		s.mu.Lock()
	}

	s.mu.Unlock()
}

func (s *Storage) ForEachCounter(fn func(key string, value int64)) {
	s.mu.Lock()

	for key, value := range s.counters {
		s.mu.Unlock()

		fn(key, value)

		s.mu.Lock()
	}

	s.mu.Unlock()
}

func (s *Storage) TotalGauges() int {
	return len(s.gauges)
}

func (s *Storage) TotalCounters() int {
	return len(s.counters)
}
