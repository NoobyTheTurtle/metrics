package collector

import (
	"context"
	"testing"
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/metric"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewGopsutilCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metrics := &metric.Metrics{}
	mockLogger := NewMockCollectorLogger(ctrl)

	collector := NewGopsutilCollector(metrics, mockLogger, 5)

	assert.NotNil(t, collector)
	assert.Equal(t, metrics, collector.metrics)
	assert.Equal(t, mockLogger, collector.logger)
	assert.Equal(t, 5*time.Second, collector.pollInterval)
}

func TestNewGopsutilCollector_ZeroPollInterval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metrics := &metric.Metrics{}
	mockLogger := NewMockCollectorLogger(ctrl)

	collector := NewGopsutilCollector(metrics, mockLogger, 0)

	assert.NotNil(t, collector)
	assert.Equal(t, metrics, collector.metrics)
	assert.Equal(t, mockLogger, collector.logger)
	assert.Equal(t, 1*time.Second, collector.pollInterval)
}

func TestGopsutilCollector_RunWithContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("context cancellation", func(t *testing.T) {
		metrics := metric.NewMetrics("localhost:8080", nil, false, "", nil)
		collector := NewGopsutilCollector(metrics, mockLogger, 1)

		mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		done := make(chan bool)
		go func() {
			collector.RunWithContext(ctx)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("RunWithContext did not stop after context cancellation")
		}
	})

	t.Run("immediate cancellation", func(t *testing.T) {
		metrics := metric.NewMetrics("localhost:8080", nil, false, "", nil)
		collector := NewGopsutilCollector(metrics, mockLogger, 10)

		mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		done := make(chan bool)
		go func() {
			collector.RunWithContext(ctx)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("RunWithContext did not stop immediately after context cancellation")
		}
	})
}

func TestGopsutilCollector_RunWithContext_ErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("with error logging", func(t *testing.T) {
		metrics := metric.NewMetrics("localhost:8080", nil, false, "", nil)
		collector := NewGopsutilCollector(metrics, mockLogger, 1)

		mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		done := make(chan bool)
		go func() {
			collector.RunWithContext(ctx)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("RunWithContext did not stop after context cancellation")
		}
	})
}

func TestGopsutilCollector_Run_NotImplemented(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockCollectorLogger(ctrl)

	metrics := metric.NewMetrics("localhost:8080", nil, false, "", nil)
	collector := NewGopsutilCollector(metrics, mockLogger, 1)

	assert.NotNil(t, collector)
}

func TestGopsutilCollector_DefaultLogicBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("covers various edge cases", func(t *testing.T) {
		metrics1 := metric.NewMetrics("", nil, false, "", nil)
		collector1 := NewGopsutilCollector(metrics1, mockLogger, 2)

		metrics2 := metric.NewMetrics("invalid:port", nil, true, "", nil)
		collector2 := NewGopsutilCollector(metrics2, mockLogger, 3)

		assert.NotNil(t, collector1)
		assert.NotNil(t, collector2)
		assert.Equal(t, 2*time.Second, collector1.pollInterval)
		assert.Equal(t, 3*time.Second, collector2.pollInterval)

		collectorMax := NewGopsutilCollector(metrics1, mockLogger, ^uint(0))
		assert.NotNil(t, collectorMax)
	})
}

func TestGopsutilCollector_RunWithContext_MultipleIterations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("multiple metric collections", func(t *testing.T) {
		metrics := metric.NewMetrics("localhost:8080", nil, false, "", nil)
		collector := NewGopsutilCollector(metrics, mockLogger, 1)

		mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

		ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
		defer cancel()

		done := make(chan bool)
		go func() {
			collector.RunWithContext(ctx)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("RunWithContext did not stop after context cancellation")
		}
	})
}

func TestGopsutilCollector_RunWithContext_ConcurrentExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger1 := NewMockCollectorLogger(ctrl)
	mockLogger2 := NewMockCollectorLogger(ctrl)

	t.Run("concurrent gopsutil collectors", func(t *testing.T) {
		metrics1 := metric.NewMetrics("localhost:8080", nil, false, "", nil)
		metrics2 := metric.NewMetrics("localhost:8081", nil, false, "", nil)

		collector1 := NewGopsutilCollector(metrics1, mockLogger1, 1)
		collector2 := NewGopsutilCollector(metrics2, mockLogger2, 1)

		mockLogger1.EXPECT().Info(gomock.Any()).AnyTimes()
		mockLogger2.EXPECT().Info(gomock.Any()).AnyTimes()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		done1 := make(chan bool)
		done2 := make(chan bool)

		go func() {
			collector1.RunWithContext(ctx)
			done1 <- true
		}()

		go func() {
			collector2.RunWithContext(ctx)
			done2 <- true
		}()

		select {
		case <-done1:
		case <-time.After(2 * time.Second):
			t.Fatal("GopsutilCollector1 did not stop after context cancellation")
		}

		select {
		case <-done2:
		case <-time.After(2 * time.Second):
			t.Fatal("GopsutilCollector2 did not stop after context cancellation")
		}
	})
}

func TestGopsutilCollector_TimingBehavior(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("fast poll interval", func(t *testing.T) {
		metrics := metric.NewMetrics("localhost:8080", nil, false, "", nil)

		collector := &GopsutilCollector{
			metrics:      metrics,
			logger:       mockLogger,
			pollInterval: 50 * time.Millisecond,
		}

		mockLogger.EXPECT().Info(gomock.Any()).MinTimes(2)

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		done := make(chan bool)
		go func() {
			collector.RunWithContext(ctx)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("RunWithContext did not stop after context cancellation")
		}
	})
}
