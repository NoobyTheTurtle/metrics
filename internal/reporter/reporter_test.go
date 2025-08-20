package reporter

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewReporter_ValidParameters(t *testing.T) {
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
}

func TestNewReporter_ZeroRateLimit_DefaultsToOne(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsReporter(ctrl)
	mockLogger := NewMockReporterLogger(ctrl)

	mockLogger.EXPECT().Info("reporter.NewReporter: Invalid rateLimit value %d provided, defaulting to 1 worker.", uint(0)).Times(1)

	reporter := NewReporter(mockMetrics, mockLogger, 5, 0)

	assert.NotNil(t, reporter)
	assert.Equal(t, uint(1), reporter.rateLimit)
}

func TestNewReporter_ZeroReportInterval_DefaultsToOne(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsReporter(ctrl)
	mockLogger := NewMockReporterLogger(ctrl)

	mockLogger.EXPECT().Info("reporter.NewReporter: Invalid reportInterval value %d provided, defaulting to 1 second.", uint(0)).Times(1)

	reporter := NewReporter(mockMetrics, mockLogger, 0, 3)

	assert.NotNil(t, reporter)
	assert.Equal(t, 1*time.Second, reporter.reportInterval)
	assert.Equal(t, uint(3), reporter.rateLimit)
}

func TestReporter_Worker_ProcessesJobsAndStopsGracefully(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsReporter(ctrl)
	mockLogger := NewMockReporterLogger(ctrl)

	reporter := NewReporter(mockMetrics, mockLogger, 5, 1)

	mockLogger.EXPECT().Info("Worker %d: started.", uint(1)).Times(1)
	mockMetrics.EXPECT().SendMetrics().Times(1)
	mockLogger.EXPECT().Info("Worker %d: successfully sent metrics.", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Worker %d: stopped.", uint(1)).Times(1)

	reporter.jobChan = make(chan struct{}, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go reporter.worker(1, &wg)

	done := make(chan struct{})
	go func() {
		defer close(done)
		reporter.jobChan <- struct{}{}
	}()

	<-done

	close(reporter.jobChan)

	wg.Wait()
}

func TestReporter_Worker_ProcessesMultipleJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsReporter(ctrl)
	mockLogger := NewMockReporterLogger(ctrl)

	reporter := NewReporter(mockMetrics, mockLogger, 5, 1)

	mockLogger.EXPECT().Info("Worker %d: started.", uint(1)).Times(1)
	mockMetrics.EXPECT().SendMetrics().Times(3)
	mockLogger.EXPECT().Info("Worker %d: successfully sent metrics.", uint(1)).Times(3)
	mockLogger.EXPECT().Info("Worker %d: stopped.", uint(1)).Times(1)

	reporter.jobChan = make(chan struct{}, 3)

	var wg sync.WaitGroup
	wg.Add(1)
	go reporter.worker(1, &wg)

	for i := 0; i < 3; i++ {
		reporter.jobChan <- struct{}{}
	}

	close(reporter.jobChan)

	wg.Wait()
}

func TestReporter_RunWithContext_GracefulShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsReporter(ctrl)
	mockLogger := NewMockReporterLogger(ctrl)

	reporter := NewReporter(mockMetrics, mockLogger, 1, 2)

	mockLogger.EXPECT().Info("Reporter starting with context. Report interval: %s. Worker pool size: %d.", "1s", uint(2)).Times(1)
	mockLogger.EXPECT().Info("Worker %d: started.", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Worker %d: started.", uint(2)).Times(1)
	mockLogger.EXPECT().Info("Reporter stopping due to context cancellation").Times(1)
	mockLogger.EXPECT().Info("Reporter waiting for workers to finish...").Times(1)
	mockLogger.EXPECT().Info("Worker %d: stopped.", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Worker %d: stopped.", uint(2)).Times(1)
	mockLogger.EXPECT().Info("Reporter stopped gracefully").Times(1)

	mockMetrics.EXPECT().SendMetrics().MaxTimes(2)
	mockLogger.EXPECT().Info("Worker %d: successfully sent metrics.", gomock.Any()).MaxTimes(2)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		reporter.RunWithContext(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RunWithContext did not stop after context cancellation")
	}
}

func TestReporter_RunWithContext_ImmediateCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsReporter(ctrl)
	mockLogger := NewMockReporterLogger(ctrl)

	reporter := NewReporter(mockMetrics, mockLogger, 10, 1)

	mockLogger.EXPECT().Info("Reporter starting with context. Report interval: %s. Worker pool size: %d.", "10s", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Worker %d: started.", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Reporter stopping due to context cancellation").Times(1)
	mockLogger.EXPECT().Info("Reporter waiting for workers to finish...").Times(1)
	mockLogger.EXPECT().Info("Worker %d: stopped.", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Reporter stopped gracefully").Times(1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		reporter.RunWithContext(ctx)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("RunWithContext did not stop immediately after context cancellation")
	}
}

func TestReporter_RunWithContext_CancellationWhileSendingJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsReporter(ctrl)
	mockLogger := NewMockReporterLogger(ctrl)

	reporter := NewReporter(mockMetrics, mockLogger, 1, 1)

	mockLogger.EXPECT().Info("Reporter starting with context. Report interval: %s. Worker pool size: %d.", "1s", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Worker %d: started.", uint(1)).Times(1)

	normalCancel := mockLogger.EXPECT().Info("Reporter stopping due to context cancellation").MaxTimes(1)
	mockLogger.EXPECT().Info("Reporter stopping due to context cancellation while sending job").MaxTimes(1)

	mockLogger.EXPECT().Info("Reporter waiting for workers to finish...").Times(1).After(normalCancel)
	mockLogger.EXPECT().Info("Worker %d: stopped.", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Reporter stopped gracefully").Times(1)

	mockMetrics.EXPECT().SendMetrics().MaxTimes(1)
	mockLogger.EXPECT().Info("Worker %d: successfully sent metrics.", uint(1)).MaxTimes(1)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		reporter.RunWithContext(ctx)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("RunWithContext did not stop after context cancellation")
	}
}

func TestReporter_RunWithContext_WorkerPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockMetricsReporter(ctrl)
	mockLogger := NewMockReporterLogger(ctrl)

	reporter := NewReporter(mockMetrics, mockLogger, 1, 1)

	mockLogger.EXPECT().Info("Reporter starting with context. Report interval: %s. Worker pool size: %d.", "1s", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Worker %d: started.", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Reporter stopping due to context cancellation").Times(1)
	mockLogger.EXPECT().Info("Reporter waiting for workers to finish...").Times(1)
	mockLogger.EXPECT().Info("Worker %d: stopped.", uint(1)).Times(1)
	mockLogger.EXPECT().Info("Reporter stopped gracefully").Times(1)

	mockMetrics.EXPECT().SendMetrics().Do(func() {
		panic("metrics sending failed")
	}).MaxTimes(1)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Caught panic in test: %v", r)
			}
		}()
		reporter.RunWithContext(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RunWithContext did not complete after worker panic")
	}
}
