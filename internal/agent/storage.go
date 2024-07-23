package agent

import "sync"

type gauges = map[string]float64
type counters = map[string]int64

// Storage describes storage that is used by agent to save collected metrics.
type Storage struct {
	mu       sync.Mutex
	gauges   gauges
	counters counters
}

// NewStorage create new storage instance.
func NewStorage() *Storage {
	return &Storage{
		gauges:   make(gauges),
		counters: make(counters),
	}
}

// UpdateGauge updates given gauge metric.
//
// name - name of metric to update.
// value - new value of gauge metric.
func (s *Storage) UpdateGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
}

// UpdateCounter updates given counter metric.
//
// name - name of metric to update.
// value - increment value that will be added to current value of counter.
func (s *Storage) UpdateCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += value
}

// ForEachGauge applies given function to every gauge metric in storage.
func (s *Storage) ForEachGauge(fn func(key string, value float64)) {
	s.mu.Lock()

	for key, value := range s.gauges {
		s.mu.Unlock()

		fn(key, value)

		s.mu.Lock()
	}

	s.mu.Unlock()
}

// ForEachCounter applies given function to every counter metric in storage.
func (s *Storage) ForEachCounter(fn func(key string, value int64)) {
	s.mu.Lock()

	for key, value := range s.counters {
		s.mu.Unlock()

		fn(key, value)

		s.mu.Lock()
	}

	s.mu.Unlock()
}

// TotalGauges returns total amount of gauge metrics in storage.
func (s *Storage) TotalGauges() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.gauges)
}

// TotalCounters returns total amount of counter metrics in storage.
func (s *Storage) TotalCounters() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.counters)
}
