package json

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUpdatesHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStorage := NewMockHandlerStorage(ctrl)
	handler := newUpdatesHandler(mockStorage)
	floatVal := 42.42
	intVal := int64(42)

	tests := []struct {
		name                 string
		body                 []byte
		updateMetricsBatch   []*model.Metric
		updateMetricsBatchFn func() *gomock.Call
		expectedStatusCode   int
		expectedBodyContains string
	}{
		{
			name: "success",
			body: []byte(`[{"id": "gauge1", "type": "gauge", "value": 42.42}, {"id": "counter1", "type": "counter", "delta": 42}]`),
			updateMetricsBatch: []*model.Metric{
				{ID: "gauge1", MType: model.GaugeType, Value: &floatVal},
				{ID: "counter1", MType: model.CounterType, Delta: &intVal},
			},
			updateMetricsBatchFn: func() *gomock.Call {
				return mockStorage.EXPECT().UpdateMetricsBatch(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid body",
			body:               []byte(`invalid`),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty metrics",
			body:               []byte(`[]`),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "empty metric id",
			body:               []byte(`[{"type": "gauge", "value": 42.42}]`),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty metric type",
			body:               []byte(`[{"id": "gauge1", "value": 42.42}]`),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "nil metric value",
			body:               []byte(`[{"id": "gauge1", "type": "gauge"}]`),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "nil metric delta",
			body:               []byte(`[{"id": "counter1", "type": "counter"}]`),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "unknown metric type",
			body:               []byte(`[{"id": "unknown1", "type": "unknown"}]`),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "error update metrics batch",
			body: []byte(`[{"id": "gauge1", "type": "gauge", "value": 42.42}]`),
			updateMetricsBatch: []*model.Metric{
				{ID: "gauge1", MType: model.GaugeType, Value: &floatVal},
			},
			updateMetricsBatchFn: func() *gomock.Call {
				return mockStorage.EXPECT().UpdateMetricsBatch(gomock.Any(), gomock.Any()).Return(errors.New("some error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tt.body))
			r = r.WithContext(context.Background())
			w := httptest.NewRecorder()

			if tt.updateMetricsBatchFn != nil {
				tt.updateMetricsBatchFn()
			}
			handler.ServeHTTP(w, r)
			assert.Equal(t, tt.expectedStatusCode, w.Code)
		})
	}
}
