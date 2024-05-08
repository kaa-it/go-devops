package service

import "context"

type Service interface {
	Ping(ctx context.Context) error
}

type Repository interface {
	Ping(ctx context.Context) error
}

type service struct {
	r Repository
}

func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) Ping(ctx context.Context) error {
	return s.r.Ping(ctx)
}
