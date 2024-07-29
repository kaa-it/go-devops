// Package api describes types for api between agent and server.
package api

// MetricsType describes type for metric type.
type MetricsType string

// Supported metric types
const (
	GaugeType   MetricsType = "gauge"   // gauge metric type
	CounterType MetricsType = "counter" // counter metric type
)

// Metrics describes one metric.
type Metrics struct {
	// ID - unique metric name.
	ID string `json:"id"`
	// MType - metric type.
	MType MetricsType `json:"type"`
	// Delta - increment value for counter metric, for gauge metric is nil.
	Delta *int64 `json:"delta,omitempty"`
	// Value - new value for gauge metric, for counter metric is nil.
	Value *float64 `json:"value,omitempty"`
}
