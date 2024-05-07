package service

type Service interface {
	Ping() error
}

type Repository interface {
	Ping() error
}

type service struct {
	r Repository
}

func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) Ping() error {
	return s.r.Ping()
}
