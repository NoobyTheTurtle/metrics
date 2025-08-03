package collector

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsCollector(ctrl)
	mockLogger := NewMockCollectorLogger(ctrl)

	collector := NewCollector(mockMetrics, mockLogger, 5)

	assert.NotNil(t, collector)
	assert.Equal(t, mockMetrics, collector.metrics)
	assert.Equal(t, mockLogger, collector.logger)
	assert.Equal(t, 5*time.Second, collector.pollInterval)
}

func TestCollector_GopsutilMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsCollector(ctrl)

	t.Run("CollectGopsutilMetrics success", func(t *testing.T) {
		mockMetrics.EXPECT().CollectGopsutilMetrics().Return(nil).Times(1)

		err := mockMetrics.CollectGopsutilMetrics()
		assert.NoError(t, err)
	})

	t.Run("CollectGopsutilMetrics error", func(t *testing.T) {
		mockMetrics.EXPECT().CollectGopsutilMetrics().Return(assert.AnError).Times(1)

		err := mockMetrics.CollectGopsutilMetrics()
		assert.Error(t, err)
	})

	t.Run("InitGomutiMetrics success", func(t *testing.T) {
		pollInterval := 5 * time.Second
		mockMetrics.EXPECT().InitGomutiMetrics(pollInterval).Return(nil).Times(1)

		err := mockMetrics.InitGomutiMetrics(pollInterval)
		assert.NoError(t, err)
	})

	t.Run("InitGomutiMetrics error", func(t *testing.T) {
		pollInterval := 5 * time.Second
		mockMetrics.EXPECT().InitGomutiMetrics(pollInterval).Return(assert.AnError).Times(1)

		err := mockMetrics.InitGomutiMetrics(pollInterval)
		assert.Error(t, err)
	})
}
