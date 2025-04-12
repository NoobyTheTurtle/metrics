package plain

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/test_utils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_handler_valueHandler(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		url                string
		setupMocks         func(*gomock.Controller) (*MockHandlerStorage, *MockHandlerLogger)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:   "successful gauge retrieval",
			method: http.MethodGet,
			url:    "/value/gauge/HeapObjects",
			setupMocks: func(ctrl *gomock.Controller) (*MockHandlerStorage, *MockHandlerLogger) {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().GetGauge("HeapObjects").Return(1.2, true)
				mockLogger := NewMockHandlerLogger(ctrl)
				return mockStorage, mockLogger
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "1.2",
		},
		{
			name:   "successful counter retrieval",
			method: http.MethodGet,
			url:    "/value/counter/PollCount",
			setupMocks: func(ctrl *gomock.Controller) (*MockHandlerStorage, *MockHandlerLogger) {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().GetCounter("PollCount").Return(int64(30), true)
				mockLogger := NewMockHandlerLogger(ctrl)
				return mockStorage, mockLogger
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "30",
		},
		{
			name:   "gauge not found",
			method: http.MethodGet,
			url:    "/value/gauge/NonExistentGauge",
			setupMocks: func(ctrl *gomock.Controller) (*MockHandlerStorage, *MockHandlerLogger) {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().GetGauge("NonExistentGauge").Return(0.0, false)
				mockLogger := NewMockHandlerLogger(ctrl)
				return mockStorage, mockLogger
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Gauge not found\n",
		},
		{
			name:   "counter not found",
			method: http.MethodGet,
			url:    "/value/counter/NonExistentCounter",
			setupMocks: func(ctrl *gomock.Controller) (*MockHandlerStorage, *MockHandlerLogger) {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().GetCounter("NonExistentCounter").Return(int64(0), false)
				mockLogger := NewMockHandlerLogger(ctrl)
				return mockStorage, mockLogger
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Counter not found\n",
		},
		{
			name:   "wrong method",
			method: http.MethodPost,
			url:    "/value/gauge/HeapObjects",
			setupMocks: func(ctrl *gomock.Controller) (*MockHandlerStorage, *MockHandlerLogger) {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockLogger := NewMockHandlerLogger(ctrl)
				return mockStorage, mockLogger
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "unknown metric type",
			method: http.MethodGet,
			url:    "/value/unknown/Metric",
			setupMocks: func(ctrl *gomock.Controller) (*MockHandlerStorage, *MockHandlerLogger) {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockLogger := NewMockHandlerLogger(ctrl)
				return mockStorage, mockLogger
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Unknown metric type\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage, logger := tt.setupMocks(ctrl)

			h := &Handler{
				storage: storage,
				logger:  logger,
			}

			r := chi.NewRouter()
			r.Get("/value/{metricType}/{metricName}", h.ValueHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, body := test_utils.TestRequest(t, ts, tt.method, tt.url)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected status code %d, got %d", tt.expectedStatusCode, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"),
					"Expected Content-Type text/plain; charset=utf-8, got %s", resp.Header.Get("Content-Type"))
			}

			assert.Equal(t, tt.expectedBody, body, "Expected response body '%s', got '%s'", tt.expectedBody, body)
		})
	}
}
