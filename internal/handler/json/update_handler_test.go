package json

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_updateHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        string
		setupMocks         func(*gomock.Controller) *MockHandlerStorage
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "successful gauge update",
			requestBody: `{
				"id": "HeapObjects",
				"type": "gauge",
				"value": 7770.0
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().UpdateGauge("HeapObjects", 7770.0).Return(7770.0, nil)
				return mockStorage
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"id":"HeapObjects","type":"gauge","value":7770}`,
		},
		{
			name: "successful counter update",
			requestBody: `{
				"id": "PollCount",
				"type": "counter",
				"delta": 30
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().UpdateCounter("PollCount", int64(30)).Return(int64(30), nil)
				return mockStorage
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"id":"PollCount","type":"counter","delta":30}`,
		},
		{
			name:        "invalid JSON format",
			requestBody: `{"id": "HeapObjects", "type": "gauge", "value": 7770.0`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				return NewMockHandlerStorage(ctrl)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid JSON format\n",
		},
		{
			name:        "missing required id and type fields",
			requestBody: `{"value": 7770.0}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				return NewMockHandlerStorage(ctrl)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "id and type fields are required\n",
		},
		{
			name: "missing required value field for gauge",
			requestBody: `{
				"id": "HeapObjects",
				"type": "gauge"
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				return NewMockHandlerStorage(ctrl)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Value field is required for gauge type\n",
		},
		{
			name: "missing required delta field for counter",
			requestBody: `{
				"id": "PollCount",
				"type": "counter"
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				return NewMockHandlerStorage(ctrl)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Delta field is required for counter type\n",
		},
		{
			name: "unknown metric type",
			requestBody: `{
				"id": "HeapObjects",
				"type": "unknown",
				"value": 7770.0
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				return NewMockHandlerStorage(ctrl)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Unknown metric type\n",
		},
		{
			name: "gauge update error",
			requestBody: `{
				"id": "HeapObjects",
				"type": "gauge",
				"value": 7770.0
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().UpdateGauge("HeapObjects", 7770.0).Return(0.0, errors.New("gauge update error"))
				return mockStorage
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Failed to update gauge\n",
		},
		{
			name: "counter update error",
			requestBody: `{
				"id": "PollCount",
				"type": "counter",
				"delta": 30
			}`,
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().UpdateCounter("PollCount", int64(30)).Return(int64(0), errors.New("counter update error"))
				return mockStorage
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Failed to update counter\n",
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
			r.Post("/update/", handler.UpdateHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, body := testutil.TestRequest(t, ts, http.MethodPost, "/update/", tt.requestBody)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected status code %d, got %d", tt.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, tt.expectedResponse, body, "Expected response body '%s', got '%s'", tt.expectedResponse, body)
		})
	}
}
