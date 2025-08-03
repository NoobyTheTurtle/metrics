package reporter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewReporter(t *testing.T) {
	t.Run("valid parameters", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMetrics := NewMockMetricsReporter(ctrl)
		mockLogger := NewMockReporterLogger(ctrl)

		reporter := NewReporter(mockMetrics, mockLogger, 5, 3)

		assert.NotNil(t, reporter)
		assert.Equal(t, mockMetrics, reporter.metrics)
		assert.Equal(t, mockLogger, reporter.logger)
		assert.Equal(t, 5*time.Second, reporter.reportInterval)
		assert.Equal(t, uint(3), reporter.rateLimit)
	})

	t.Run("zero rate limit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMetrics := NewMockMetricsReporter(ctrl)
		mockLogger := NewMockReporterLogger(ctrl)

		mockLogger.EXPECT().Info("reporter.NewReporter: Invalid rateLimit value %d provided, defaulting to 1 worker.", uint(0)).Times(1)

		reporter := NewReporter(mockMetrics, mockLogger, 5, 0)

		assert.NotNil(t, reporter)
		assert.Equal(t, uint(1), reporter.rateLimit)
	})
}

func TestReporter_Worker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsReporter(ctrl)
	mockLogger := NewMockReporterLogger(ctrl)

	reporter := NewReporter(mockMetrics, mockLogger, 5, 1)

	mockLogger.EXPECT().Info("Worker %d: started.", uint(1)).Times(1)
	mockMetrics.EXPECT().SendMetrics().Times(1)
	mockLogger.EXPECT().Info("Worker %d: successfully sent metrics.", uint(1)).Times(1)

	reporter.jobChan = make(chan struct{}, 1)
	defer close(reporter.jobChan)

	go reporter.worker(1)

	reporter.jobChan <- struct{}{}

	time.Sleep(100 * time.Millisecond)
}
