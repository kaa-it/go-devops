package updating

type Service interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	Gauge(name string) (float64, error)
	Counter(name string) (int64, error)
}

type Repository interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	Gauge(name string) (float64, error)
	Counter(name string) (int64, error)
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

func (s *service) Gauge(name string) (float64, error) {
	return s.r.Gauge(name)
}

func (s *service) Counter(name string) (int64, error) {
	return s.r.Counter(name)
}
