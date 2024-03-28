package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_UpdateGauge(t *testing.T) {
	t.Run("update once", func(t *testing.T) {
		s := NewStorage()

		s.UpdateGauge("TestMetric", 6.04)

		value, ok := s.gauges["TestMetric"]

		assert.True(t, ok)
		assert.Equal(t, value, 6.04)
	})

	t.Run("update twice", func(t *testing.T) {
		s := NewStorage()

		s.UpdateGauge("TestMetric", 6.04)
		s.UpdateGauge("TestMetric", 7.04)

		value, ok := s.gauges["TestMetric"]

		assert.True(t, ok)
		assert.Equal(t, value, 7.04)
	})
}

func TestRepository_UpdateCounter(t *testing.T) {
	t.Run("update once", func(t *testing.T) {
		s := NewStorage()

		s.UpdateCounter("TestMetric", 60)

		value, ok := s.counters["TestMetric"]

		assert.True(t, ok)
		assert.Equal(t, value, int64(60))
	})

	t.Run("update twice", func(t *testing.T) {
		s := NewStorage()

		s.UpdateCounter("TestMetric", 60)
		s.UpdateCounter("TestMetric", 60)

		value, ok := s.counters["TestMetric"]

		assert.True(t, ok)
		assert.Equal(t, value, int64(120))
	})
}

func TestRepository_ForEachGauge(t *testing.T) {
	s := NewStorage()

	var called int

	mockFn := func(key string, value float64) {
		called += 1
	}

	s.UpdateGauge("TestMetric1", 6.04)
	s.UpdateGauge("TestMetric2", 7.04)
	s.UpdateCounter("TestMetric3", 60)
	s.UpdateCounter("TestMetric3", 60)

	s.ForEachGauge(mockFn)

	assert.Equal(t, called, 2)
}

func TestRepository_ForEachCounter(t *testing.T) {
	s := NewStorage()

	var called int

	mockFn := func(key string, value int64) {
		called += 1
	}

	s.UpdateGauge("TestMetric1", 6.04)
	s.UpdateGauge("TestMetric2", 7.04)
	s.UpdateCounter("TestMetric3", 60)
	s.UpdateCounter("TestMetric3", 60)

	s.ForEachCounter(mockFn)

	assert.Equal(t, called, 1)
}
