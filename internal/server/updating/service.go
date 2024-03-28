package updating

type Service interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

type Repository interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

type service struct {
	r Repository
}

func NewService(r Repository) Service {
	return &service{r}
}

func (s *service) UpdateGauge(name string, value float64) {
	s.r.UpdateGauge(name, value)
}

func (s *service) UpdateCounter(name string, value int64) {
	s.r.UpdateCounter(name, value)
}
