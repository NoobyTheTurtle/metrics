package json

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUpdatesHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockHandlerStorage(ctrl)
	handler := newUpdatesHandler(mockStorage)

	body := []byte(`[{"id": "gauge1", "type": "gauge", "value": 42.42}, {"id": "counter1", "type": "counter", "delta": 42}]`)

	mockStorage.EXPECT().UpdateMetricsBatch(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdatesHandler_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockHandlerStorage(ctrl)
	handler := newUpdatesHandler(mockStorage)

	body := []byte(`invalid json`)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatesHandler_EmptyMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockHandlerStorage(ctrl)
	handler := newUpdatesHandler(mockStorage)

	body := []byte(`[]`)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdatesHandler_ValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected int
	}{
		{
			name:     "empty metric id",
			body:     `[{"type": "gauge", "value": 42.42}]`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "empty metric type",
			body:     `[{"id": "gauge1", "value": 42.42}]`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "nil gauge value",
			body:     `[{"id": "gauge1", "type": "gauge"}]`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "nil counter delta",
			body:     `[{"id": "counter1", "type": "counter"}]`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "unknown metric type",
			body:     `[{"id": "unknown1", "type": "unknown"}]`,
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockHandlerStorage(ctrl)
			handler := newUpdatesHandler(mockStorage)

			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.body)))
			r = r.WithContext(context.Background())
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestUpdatesHandler_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockHandlerStorage(ctrl)
	handler := newUpdatesHandler(mockStorage)

	body := []byte(`[{"id": "gauge1", "type": "gauge", "value": 42.42}]`)
	mockStorage.EXPECT().UpdateMetricsBatch(gomock.Any(), gomock.Any()).Return(errors.New("storage error")).Times(1)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdatesHandler_LargePayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockHandlerStorage(ctrl)
	handler := newUpdatesHandler(mockStorage)

	var metrics []string
	for i := 0; i < 100; i++ {
		metrics = append(metrics, fmt.Sprintf(`{"id": "metric_%d", "type": "gauge", "value": %d.5}`, i, i))
	}
	body := []byte("[" + strings.Join(metrics, ",") + "]")

	mockStorage.EXPECT().UpdateMetricsBatch(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdatesHandler_MixedMetricTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockHandlerStorage(ctrl)
	handler := newUpdatesHandler(mockStorage)

	body := []byte(`[
		{"id": "gauge1", "type": "gauge", "value": 10.5},
		{"id": "counter1", "type": "counter", "delta": 5},
		{"id": "gauge2", "type": "gauge", "value": 42.42},
		{"id": "counter2", "type": "counter", "delta": 10}
	]`)

	mockStorage.EXPECT().UpdateMetricsBatch(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdatesHandler_ZeroValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockHandlerStorage(ctrl)
	handler := newUpdatesHandler(mockStorage)

	body := []byte(`[
		{"id": "gauge_zero", "type": "gauge", "value": 0.0},
		{"id": "counter_zero", "type": "counter", "delta": 0}
	]`)

	mockStorage.EXPECT().UpdateMetricsBatch(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdatesHandler_ExtremeValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockHandlerStorage(ctrl)
	handler := newUpdatesHandler(mockStorage)

	body := []byte(fmt.Sprintf(`[
		{"id": "gauge_max", "type": "gauge", "value": %g},
		{"id": "gauge_min", "type": "gauge", "value": %g},
		{"id": "counter_max", "type": "counter", "delta": %d}
	]`, math.MaxFloat64, -math.MaxFloat64, math.MaxInt64))

	mockStorage.EXPECT().UpdateMetricsBatch(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}
