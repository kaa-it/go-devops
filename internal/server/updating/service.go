package updating

import (
	"context"
	"github.com/kaa-it/go-devops/internal/api"
)

type Service interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	Gauge(ctx context.Context, name string) (float64, error)
	Counter(ctx context.Context, name string) (int64, error)
	Updates(ctx context.Context, metrics []api.Metrics) error
}

type Repository interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	Gauge(ctx context.Context, name string) (float64, error)
	Counter(ctx context.Context, name string) (int64, error)
	Updates(ctx context.Context, metrics []api.Metrics) error
}

type service struct {
	r Repository
}

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
