package memory

import (
	"context"
	"github.com/kaa-it/go-devops/internal/api"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_Updates(t *testing.T) {
	t.Run("gauge and counter", func(t *testing.T) {
		s, err := NewStorage(&StorageConfig{
			StoreInterval: 0,
			StoreFilePath: "",
			Restore:       false,
		})

		assert.NoError(t, err)

		var gauge = 6.7
		var delta1 int64 = 16
		var delta2 int64 = 10

		metrics := []api.Metrics{
			{
				ID:    "TestGauge",
				MType: api.GaugeType,
				Delta: nil,
				Value: &gauge,
			},
			{
				ID:    "TestCounter",
				MType: api.CounterType,
				Delta: &delta1,
				Value: nil,
			},
			{
				ID:    "TestCounter",
				MType: api.CounterType,
				Delta: &delta2,
				Value: nil,
			},
		}

		err = s.Updates(context.Background(), metrics)

		assert.NoError(t, err)

		gaugeMetric, err := s.Gauge(context.Background(), "TestGauge")

		assert.NoError(t, err)
		assert.Equal(t, gauge, gaugeMetric)

		counterMetric, err := s.Counter(context.Background(), "TestCounter")

		assert.NoError(t, err)
		assert.Equal(t, delta1+delta2, counterMetric)

		gauges, err := s.TotalGauges(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 1, gauges)

		counters, err := s.TotalCounters(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 1, counters)
	})
}

func TestRepository_UpdateGauge(t *testing.T) {
	t.Run("update once", func(t *testing.T) {
		s, err := NewStorage(&StorageConfig{
			StoreInterval: 0,
			StoreFilePath: "",
			Restore:       false,
		})

		assert.NoError(t, err)

		err = s.UpdateGauge(context.Background(), "TestMetric", 6.04)

		assert.NoError(t, err)

		value, ok := s.gauges["TestMetric"]

		assert.True(t, ok)
		assert.Equal(t, value, 6.04)
	})

	t.Run("update twice", func(t *testing.T) {
		s, err := NewStorage(&StorageConfig{
			StoreInterval: 0,
			StoreFilePath: "",
			Restore:       false,
		})

		assert.NoError(t, err)

		err = s.UpdateGauge(context.Background(), "TestMetric", 6.04)

		assert.NoError(t, err)

		err = s.UpdateGauge(context.Background(), "TestMetric", 7.04)

		assert.NoError(t, err)

		value, ok := s.gauges["TestMetric"]

		assert.True(t, ok)
		assert.Equal(t, value, 7.04)
	})
}

func TestRepository_UpdateCounter(t *testing.T) {
	t.Run("update once", func(t *testing.T) {
		s, err := NewStorage(&StorageConfig{
			StoreInterval: 0,
			StoreFilePath: "",
			Restore:       false,
		})

		assert.NoError(t, err)

		err = s.UpdateCounter(context.Background(), "TestMetric", 60)

		assert.NoError(t, err)

		value, ok := s.counters["TestMetric"]

		assert.True(t, ok)
		assert.Equal(t, value, int64(60))
	})

	t.Run("update twice", func(t *testing.T) {
		s, err := NewStorage(&StorageConfig{
			StoreInterval: 0,
			StoreFilePath: "",
			Restore:       false,
		})

		assert.NoError(t, err)

		err = s.UpdateCounter(context.Background(), "TestMetric", 60)

		assert.NoError(t, err)

		err = s.UpdateCounter(context.Background(), "TestMetric", 60)

		assert.NoError(t, err)

		value, ok := s.counters["TestMetric"]

		assert.True(t, ok)
		assert.Equal(t, value, int64(120))
	})
}

func TestRepository_ForEachGauge(t *testing.T) {
	s, err := NewStorage(&StorageConfig{
		StoreInterval: 0,
		StoreFilePath: "",
		Restore:       false,
	})

	assert.NoError(t, err)

	var called int

	mockFn := func(key string, value float64) {
		called += 1
	}

	err = s.UpdateGauge(context.Background(), "TestMetric1", 6.04)

	assert.NoError(t, err)

	err = s.UpdateGauge(context.Background(), "TestMetric2", 7.04)

	assert.NoError(t, err)

	err = s.UpdateCounter(context.Background(), "TestMetric3", 60)

	assert.NoError(t, err)

	err = s.UpdateCounter(context.Background(), "TestMetric3", 60)

	assert.NoError(t, err)

	err = s.ForEachGauge(context.Background(), mockFn)

	assert.NoError(t, err)
	assert.Equal(t, called, 2)
}

func TestRepository_ForEachCounter(t *testing.T) {
	s, err := NewStorage(&StorageConfig{
		StoreInterval: 0,
		StoreFilePath: "",
		Restore:       false,
	})

	assert.NoError(t, err)

	var called int

	mockFn := func(key string, value int64) {
		called += 1
	}

	err = s.UpdateGauge(context.Background(), "TestMetric1", 6.04)

	assert.NoError(t, err)

	err = s.UpdateGauge(context.Background(), "TestMetric2", 7.04)

	assert.NoError(t, err)

	err = s.UpdateCounter(context.Background(), "TestMetric3", 60)

	assert.NoError(t, err)

	err = s.UpdateCounter(context.Background(), "TestMetric3", 60)

	assert.NoError(t, err)

	err = s.ForEachCounter(context.Background(), mockFn)
	assert.NoError(t, err)

	assert.Equal(t, called, 1)
}
