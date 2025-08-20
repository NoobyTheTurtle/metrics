package persister

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewPersister_ValidParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockMetricsStorage(ctrl)
	mockLogger := NewMockPersisterLogger(ctrl)

	persister := NewPersister(mockStorage, mockLogger, 5)

	assert.NotNil(t, persister)
	assert.Equal(t, mockStorage, persister.storage)
	assert.Equal(t, mockLogger, persister.logger)
	assert.Equal(t, 5*time.Second, persister.interval)
}

func TestNewPersister_ZeroInterval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockMetricsStorage(ctrl)
	mockLogger := NewMockPersisterLogger(ctrl)

	persister := NewPersister(mockStorage, mockLogger, 0)

	assert.NotNil(t, persister)
	assert.Equal(t, 0*time.Second, persister.interval)
}

func TestPersister_Run_GracefulShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockMetricsStorage(ctrl)
	mockLogger := NewMockPersisterLogger(ctrl)

	mockLogger.EXPECT().Info("Periodic saving enabled with interval %v", 100*time.Millisecond).Times(1)
	mockLogger.EXPECT().Info("Persister shutting down, performing final save...").Times(1)
	mockLogger.EXPECT().Info("Final save completed successfully").Times(1)

	mockStorage.EXPECT().SaveToFile(gomock.Any()).Return(nil).Times(1)

	persister := NewPersister(mockStorage, mockLogger, 0)
	persister.interval = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		persister.Run(ctx)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Persister did not shutdown gracefully")
	}
}

func TestPersister_Run_PeriodicSave(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockMetricsStorage(ctrl)
	mockLogger := NewMockPersisterLogger(ctrl)

	mockLogger.EXPECT().Info("Periodic saving enabled with interval %v", 50*time.Millisecond).Times(1)
	mockLogger.EXPECT().Info("Successfully saved metrics to file").MinTimes(1).MaxTimes(3)
	mockLogger.EXPECT().Info("Persister shutting down, performing final save...").Times(1)
	mockLogger.EXPECT().Info("Final save completed successfully").Times(1)

	mockStorage.EXPECT().SaveToFile(gomock.Any()).Return(nil).MinTimes(2).MaxTimes(4)

	persister := NewPersister(mockStorage, mockLogger, 0)
	persister.interval = 50 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		persister.Run(ctx)
	}()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Persister did not shutdown gracefully")
	}
}

func TestPersister_Run_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockMetricsStorage(ctrl)
	mockLogger := NewMockPersisterLogger(ctrl)

	saveErr := errors.New("failed to save file")

	mockLogger.EXPECT().Info("Periodic saving enabled with interval %v", 50*time.Millisecond).Times(1)
	mockLogger.EXPECT().Error("Failed to save metrics: %v", saveErr).MinTimes(1)
	mockLogger.EXPECT().Info("Persister shutting down, performing final save...").Times(1)
	mockLogger.EXPECT().Error("Failed to perform final save during shutdown: %v", saveErr).Times(1)

	mockStorage.EXPECT().SaveToFile(gomock.Any()).Return(saveErr).MinTimes(2)

	persister := NewPersister(mockStorage, mockLogger, 0)
	persister.interval = 50 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		persister.Run(ctx)
	}()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Persister did not handle save errors gracefully")
	}
}

func TestPersister_Run_ImmediateCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockMetricsStorage(ctrl)
	mockLogger := NewMockPersisterLogger(ctrl)

	mockLogger.EXPECT().Info("Periodic saving enabled with interval %v", 100*time.Millisecond).Times(1)
	mockLogger.EXPECT().Info("Persister shutting down, performing final save...").Times(1)
	mockLogger.EXPECT().Info("Final save completed successfully").Times(1)

	mockStorage.EXPECT().SaveToFile(gomock.Any()).Return(nil).Times(1)

	persister := NewPersister(mockStorage, mockLogger, 0)
	persister.interval = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		persister.Run(ctx)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Persister did not handle immediate cancellation gracefully")
	}
}

func TestPersister_Run_FinalSaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockMetricsStorage(ctrl)
	mockLogger := NewMockPersisterLogger(ctrl)

	finalSaveErr := errors.New("final save failed")

	mockLogger.EXPECT().Info("Periodic saving enabled with interval %v", 50*time.Millisecond).Times(1)
	mockLogger.EXPECT().Info("Successfully saved metrics to file").MaxTimes(2)
	mockLogger.EXPECT().Info("Persister shutting down, performing final save...").Times(1)
	mockLogger.EXPECT().Error("Failed to perform final save during shutdown: %v", finalSaveErr).Times(1)

	mockStorage.EXPECT().SaveToFile(gomock.Not(gomock.Eq(context.Background()))).Return(nil).MaxTimes(2)
	mockStorage.EXPECT().SaveToFile(context.Background()).Return(finalSaveErr).Times(1)

	persister := NewPersister(mockStorage, mockLogger, 0)
	persister.interval = 50 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		persister.Run(ctx)
	}()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Persister did not handle final save error gracefully")
	}
}
