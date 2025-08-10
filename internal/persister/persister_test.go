package persister

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewPersister(t *testing.T) {
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
