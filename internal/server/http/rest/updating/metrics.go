package updating

type MetricsType string

const (
	GaugeType   MetricsType = "gauge"
	CounterType MetricsType = "counter"
)

type Metrics struct {
	ID    string      `json:"id"`
	MType MetricsType `json:"type"`
	Delta *int64      `json:"delta,omitempty"`
	Value *float64    `json:"value,omitempty"`
}
