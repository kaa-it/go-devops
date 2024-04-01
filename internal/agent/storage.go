package agent

import "sync"

type gauges = map[string]float64
type counters = map[string]int64

type Storage struct {
	mu       sync.Mutex
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
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.gauges)
}

func (s *Storage) TotalCounters() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.counters)
}
