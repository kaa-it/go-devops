// Package service describes service for testing storage connection.
package service

import "context"

// Service describes methods provided by service.
type Service interface {
	// Ping tests storage connection.
	Ping(ctx context.Context) error
}

// Repository describes methods for repository that must be provided to the service.
type Repository interface {
	// Ping tests storage connection.
	Ping(ctx context.Context) error
}

type service struct {
	r Repository
}

// NewService creates new service instance.
func NewService(r Repository) Service {
	return &service{r}
}

// Ping checks storage connection.
func (s *service) Ping(ctx context.Context) error {
	return s.r.Ping(ctx)
}
