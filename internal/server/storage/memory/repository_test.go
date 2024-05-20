package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	s.ForEachGauge(context.Background(), mockFn)

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

	s.ForEachCounter(context.Background(), mockFn)

	assert.Equal(t, called, 1)
}
