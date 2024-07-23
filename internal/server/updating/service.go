// Package updating provides service for updating context boundary.
package updating

import (
	"context"

	"github.com/kaa-it/go-devops/internal/api"
)

// Service describes methods provided by the service.
type Service interface {
	// UpdateGauge updates gauge metric with given name and value.
	UpdateGauge(ctx context.Context, name string, value float64) error
	// UpdateCounter updates counter metric with given name and value.
	UpdateCounter(ctx context.Context, name string, value int64) error
	// Gauge returns gauge metric value with given name.
	Gauge(ctx context.Context, name string) (float64, error)
	// Counter returns counter metric value with given name.
	Counter(ctx context.Context, name string) (int64, error)
	// Updates updates some metrics simultaneously.
	Updates(ctx context.Context, metrics []api.Metrics) error
}

// Repository describes methods for repository that must be provided to the service.
// The service uses this repository to update metrics in storage.
type Repository interface {
	// UpdateGauge updates gauge metric in storage with given name and value.
	UpdateGauge(ctx context.Context, name string, value float64) error
	// UpdateCounter updates counter metric in storage with given name and value.
	UpdateCounter(ctx context.Context, name string, value int64) error
	// Gauge returns gauge metric value with given name from storage.
	Gauge(ctx context.Context, name string) (float64, error)
	// Counter returns counter metric value with given name from storage.
	Counter(ctx context.Context, name string) (int64, error)
	// Updates updates some metrics simultaneously in storage.
	Updates(ctx context.Context, metrics []api.Metrics) error
}

type service struct {
	r Repository
}

// NewService creates new service instance.
func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) UpdateGauge(ctx context.Context, name string, value float64) error {
	return s.r.UpdateGauge(ctx, name, value)
}

func (s *service) UpdateCounter(ctx context.Context, name string, value int64) error {
	return s.r.UpdateCounter(ctx, name, value)
}

func (s *service) Gauge(ctx context.Context, name string) (float64, error) {
	return s.r.Gauge(ctx, name)
}

func (s *service) Counter(ctx context.Context, name string) (int64, error) {
	return s.r.Counter(ctx, name)
}

func (s *service) Updates(ctx context.Context, metrics []api.Metrics) error {
	return s.r.Updates(ctx, metrics)
}
