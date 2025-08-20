package collector

import (
	"context"
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

func TestNewCollector_ZeroPollInterval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsCollector(ctrl)
	mockLogger := NewMockCollectorLogger(ctrl)

	collector := NewCollector(mockMetrics, mockLogger, 0)

	assert.NotNil(t, collector)
	assert.Equal(t, mockMetrics, collector.metrics)
	assert.Equal(t, mockLogger, collector.logger)
	assert.Equal(t, 1*time.Second, collector.pollInterval)
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

	t.Run("InitGopsutilMetrics success", func(t *testing.T) {
		pollInterval := 5 * time.Second
		mockMetrics.EXPECT().InitGopsutilMetrics(pollInterval).Return(nil).Times(1)

		err := mockMetrics.InitGopsutilMetrics(pollInterval)
		assert.NoError(t, err)
	})

	t.Run("InitGopsutilMetrics error", func(t *testing.T) {
		pollInterval := 5 * time.Second
		mockMetrics.EXPECT().InitGopsutilMetrics(pollInterval).Return(assert.AnError).Times(1)

		err := mockMetrics.InitGopsutilMetrics(pollInterval)
		assert.Error(t, err)
	})
}

func TestCollector_RunWithContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsCollector(ctrl)
	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("context cancellation", func(t *testing.T) {
		collector := NewCollector(mockMetrics, mockLogger, 1)

		mockMetrics.EXPECT().UpdateMetrics().AnyTimes()
		mockLogger.EXPECT().Info("Metrics updated").AnyTimes()
		mockLogger.EXPECT().Info("Collector stopping due to context cancellation").Times(1)

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
		collector := NewCollector(mockMetrics, mockLogger, 10)

		mockLogger.EXPECT().Info("Collector stopping due to context cancellation").Times(1)

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

func TestCollector_Run_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsCollector(ctrl)
	mockLogger := NewMockCollectorLogger(ctrl)

	collector := NewCollector(mockMetrics, mockLogger, 1)

	mockMetrics.EXPECT().UpdateMetrics().AnyTimes()
	mockLogger.EXPECT().Info("Metrics updated").AnyTimes()
	mockLogger.EXPECT().Info("Collector stopping due to context cancellation").Times(1)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	done := make(chan bool)
	go func() {
		defer func() { done <- true }()

		collector.RunWithContext(ctx)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Collector did not stop within reasonable time")
	}
}

func TestCollector_RunWithContext_MultipleIterations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsCollector(ctrl)
	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("multiple update calls", func(t *testing.T) {
		collector := NewCollector(mockMetrics, mockLogger, 1)

		mockMetrics.EXPECT().UpdateMetrics().Times(2)
		mockLogger.EXPECT().Info("Metrics updated").Times(2)
		mockLogger.EXPECT().Info("Collector stopping due to context cancellation").Times(1)

		ctx, cancel := context.WithTimeout(context.Background(), 2100*time.Millisecond)
		defer cancel()

		done := make(chan bool)
		go func() {
			collector.RunWithContext(ctx)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatal("RunWithContext did not stop after context cancellation")
		}
	})
}

func TestCollector_RunWithContext_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsCollector(ctrl)
	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("concurrent collectors", func(t *testing.T) {
		collector1 := NewCollector(mockMetrics, mockLogger, 1)
		collector2 := NewCollector(mockMetrics, mockLogger, 1)

		mockMetrics.EXPECT().UpdateMetrics().AnyTimes()
		mockLogger.EXPECT().Info("Metrics updated").AnyTimes()
		mockLogger.EXPECT().Info("Collector stopping due to context cancellation").Times(2)

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
			t.Fatal("Collector1 did not stop after context cancellation")
		}

		select {
		case <-done2:
		case <-time.After(2 * time.Second):
			t.Fatal("Collector2 did not stop after context cancellation")
		}
	})
}

func TestCollector_EdgeCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsCollector(ctrl)
	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("edge case poll intervals", func(t *testing.T) {
		collectorMax := NewCollector(mockMetrics, mockLogger, ^uint(0))
		assert.NotNil(t, collectorMax)

		collector1 := NewCollector(mockMetrics, mockLogger, 1)
		assert.Equal(t, 1*time.Second, collector1.pollInterval)

		collectorDirect := &Collector{
			metrics:      mockMetrics,
			logger:       mockLogger,
			pollInterval: 500 * time.Millisecond,
		}
		assert.NotNil(t, collectorDirect)
		assert.Equal(t, 500*time.Millisecond, collectorDirect.pollInterval)
	})
}

func TestCollector_RunWithContext_EdgeCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsCollector(ctrl)
	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("pre-cancelled context", func(t *testing.T) {
		collector := NewCollector(mockMetrics, mockLogger, 5)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockLogger.EXPECT().Info("Collector stopping due to context cancellation").Times(1)

		done := make(chan bool)
		go func() {
			collector.RunWithContext(ctx)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("RunWithContext should return immediately with cancelled context")
		}
	})
}

func TestCollector_TimingBehavior(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsCollector(ctrl)
	mockLogger := NewMockCollectorLogger(ctrl)

	t.Run("fast poll interval", func(t *testing.T) {
		collector := &Collector{
			metrics:      mockMetrics,
			logger:       mockLogger,
			pollInterval: 50 * time.Millisecond,
		}

		mockMetrics.EXPECT().UpdateMetrics().MinTimes(2)
		mockLogger.EXPECT().Info("Metrics updated").MinTimes(2)
		mockLogger.EXPECT().Info("Collector stopping due to context cancellation").Times(1)

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
