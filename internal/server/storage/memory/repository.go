package memory

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	fileName string
}

func NewStorage() *Storage {
	return &Storage{
		gauges:   make(gauges),
		counters: make(counters),
		fileName: "",
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

	file, err := os.OpenFile(s.fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	for name, value := range s.gauges {
		str := fmt.Sprintf("%s\t%s\t%f\n", "gauge", name, value)
		if _, err := file.WriteString(str); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) Load() error {
	file, err := os.Open(s.fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")

		if len(parts) != 3 {
			return err
		}

		switch parts[0] {
		case "counter":
			value, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return err
			}

			s.counters[parts[1]] = value
		case "gauge":
			value, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				return err
			}

			s.gauges[parts[1]] = value
		}
	}

	return nil
}

// TODO: Запускаем горутину для сохраниения, читаем при старте; останов горутины; запись при завершении приложения;
// TODO: корректное сохранение горутины; запись всего синхронная, если период 0
