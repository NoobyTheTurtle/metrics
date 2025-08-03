package collector

import (
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
