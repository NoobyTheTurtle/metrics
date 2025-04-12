package json

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_valueHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        string
		setupMocks         func(*gomock.Controller) *MockHandlerStorage
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "successful gauge retrieval",
			requestBody: `{
				"id": "HeapObjects",
				"type": "gauge"
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().GetGauge("HeapObjects").Return(7770.0, true)
				return mockStorage
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"id":"HeapObjects","type":"gauge","value":7770}`,
		},
		{
			name: "successful counter retrieval",
			requestBody: `{
				"id": "PollCount",
				"type": "counter"
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().GetCounter("PollCount").Return(int64(30), true)
				return mockStorage
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"id":"PollCount","type":"counter","delta":30}`,
		},
		{
			name:        "invalid JSON format",
			requestBody: `{"id": "HeapObjects", "type": "gauge"`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				return NewMockHandlerStorage(ctrl)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid JSON format\n",
		},
		{
			name:        "missing required id and type fields",
			requestBody: `{}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				return NewMockHandlerStorage(ctrl)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "id and type fields are required\n",
		},
		{
			name: "gauge metric not found",
			requestBody: `{
				"id": "NonExistentGauge",
				"type": "gauge"
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().GetGauge("NonExistentGauge").Return(0.0, false)
				return mockStorage
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   "Gauge not found\n",
		},
		{
			name: "counter metric not found",
			requestBody: `{
				"id": "NonExistentCounter",
				"type": "counter"
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().GetCounter("NonExistentCounter").Return(int64(0), false)
				return mockStorage
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   "Counter not found\n",
		},
		{
			name: "unknown metric type",
			requestBody: `{
				"id": "HeapObjects",
				"type": "unknown"
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				return NewMockHandlerStorage(ctrl)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Unknown metric type\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := tt.setupMocks(ctrl)
			handler := &Handler{
				storage: storage,
			}

			r := chi.NewRouter()
			r.Post("/value/", handler.ValueHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, body := testutil.TestRequest(t, ts, http.MethodPost, "/value/", tt.requestBody)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected status code %d, got %d", tt.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, tt.expectedResponse, body, "Expected response body '%s', got '%s'", tt.expectedResponse, body)
		})
	}
}
