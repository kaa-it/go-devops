package viewing

import "fmt"

type Service interface {
	Gauge(name string) (float64, error)
	Counter(name string) (int64, error)
	Gauges() ([]Gauge, error)
	Counters() ([]Counter, error)
}

type Repository interface {
	Gauge(name string) (float64, error)
	Counter(name string) (int64, error)
	ForEachGauge(fn func(key string, value float64))
	ForEachCounter(fn func(key string, value int64))
	TotalGauges() int
	TotalCounters() int
}

type service struct {
	r Repository
}

func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) Gauge(name string) (float64, error) {
	value, err := s.r.Gauge(name)
	if err != nil {
		return 0, fmt.Errorf("failed to get %s gauge: %w", name, err)
	}

	return value, nil
}

func (s *service) Counter(name string) (int64, error) {
	value, err := s.r.Counter(name)
	if err != nil {
		return 0, fmt.Errorf("failed to get %s counter: %w", name, err)
	}

	return value, nil
}

func (s *service) Gauges() ([]Gauge, error) {
	gauges := make([]Gauge, 0, s.r.TotalGauges())

	s.r.ForEachGauge(func(key string, value float64) {
		gauges = append(gauges, Gauge{
			Name:  key,
			Value: value,
		})
	})

	return gauges, nil
}

func (s *service) Counters() ([]Counter, error) {
	counters := make([]Counter, 0, s.r.TotalCounters())

	s.r.ForEachCounter(func(key string, value int64) {
		counters = append(counters, Counter{
			Name:  key,
			Value: value,
		})
	})

	return counters, nil
}
