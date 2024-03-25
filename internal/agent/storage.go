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
