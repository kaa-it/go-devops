// Package viewing provides service for viewing context boundary.
package viewing

import (
	"context"
	"fmt"
)

// Service describes methods provided by the service.
type Service interface {
	// Gauge returns value for gauge metric with given name.
	Gauge(ctx context.Context, name string) (float64, error)
	// Counter returns value for counter metrics with given name.
	Counter(ctx context.Context, name string) (int64, error)
	// Gauges returns all gauge metrics.
	Gauges(ctx context.Context) ([]Gauge, error)
	// Counters returns all counter metrics.
	Counters(ctx context.Context) ([]Counter, error)
}

// Repository describes methods for repository that must be provided to the service.
// The service uses this repository to get metrics from storage.
type Repository interface {
	// Gauge returns gauge metric value with given name.
	Gauge(ctx context.Context, name string) (float64, error)
	// Counter returns counter metric value with given name.
	Counter(ctx context.Context, name string) (int64, error)
	// ForEachGauge applies given function to every gauge metric in storage.
	ForEachGauge(ctx context.Context, fn func(key string, value float64)) error
	// ForEachCounter applies given function to every counter metric in storage.
	ForEachCounter(ctx context.Context, fn func(key string, value int64)) error
	// TotalGauges returns total amount of gauge metrics in storage.
	TotalGauges(ctx context.Context) (int, error)
	// TotalCounters returns total amount of counter metrics in storage.
	TotalCounters(ctx context.Context) (int, error)
}

type service struct {
	r Repository
}

// NewService creates new service instance.
func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) Gauge(ctx context.Context, name string) (float64, error) {
	value, err := s.r.Gauge(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("failed to get %s gauge: %w", name, err)
	}

	return value, nil
}

func (s *service) Counter(ctx context.Context, name string) (int64, error) {
	value, err := s.r.Counter(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("failed to get %s counter: %w", name, err)
	}

	return value, nil
}

func (s *service) Gauges(ctx context.Context) ([]Gauge, error) {
	total, err := s.r.TotalGauges(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total gauges: %w", err)
	}

	gauges := make([]Gauge, 0, total)

	err = s.r.ForEachGauge(ctx, func(key string, value float64) {
		gauges = append(gauges, Gauge{
			Name:  key,
			Value: value,
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get gauges: %w", err)
	}

	return gauges, nil
}

func (s *service) Counters(ctx context.Context) ([]Counter, error) {
	total, err := s.r.TotalCounters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total counters: %w", err)
	}

	counters := make([]Counter, 0, total)

	err = s.r.ForEachCounter(ctx, func(key string, value int64) {
		counters = append(counters, Counter{
			Name:  key,
			Value: value,
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get counters: %w", err)
	}

	return counters, nil
}
