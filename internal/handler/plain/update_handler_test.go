package plain

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

func Test_handler_updateHandler(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		url                string
		setupMocks         func(*gomock.Controller) *MockHandlerStorage
		expectedStatusCode int
	}{
		{
			name:   "successful gauge update",
			method: http.MethodPost,
			url:    "/update/gauge/HeapObjects/7770",
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().UpdateGauge("HeapObjects", 7770.0).Return(7770.0, nil)

				return mockStorage
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:   "successful counter update",
			method: http.MethodPost,
			url:    "/update/counter/PollCount/30",
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().UpdateCounter("PollCount", int64(30)).Return(int64(30), nil)

				return mockStorage
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:   "wrong method",
			method: http.MethodGet,
			url:    "/update/counter/PollCount/30",
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)

				return mockStorage
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "invalid url format",
			method: http.MethodPost,
			url:    "/update/gauge/wrong-format",
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)

				return mockStorage
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:   "empty metric name",
			method: http.MethodPost,
			url:    "/update/gauge//30",
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)

				return mockStorage
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:   "invalid gauge value",
			method: http.MethodPost,
			url:    "/update/gauge/HeapObjects/not-a-number",
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)

				return mockStorage
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "invalid counter value",
			method: http.MethodPost,
			url:    "/update/counter/PollCount/not-a-number",
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)

				return mockStorage
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "unknown metric type",
			method: http.MethodPost,
			url:    "/update/unknown/PollCount/30",
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)

				return mockStorage
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "gauge update error",
			method: http.MethodPost,
			url:    "/update/gauge/HeapObjects/7770",
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().UpdateGauge("HeapObjects", 7770.0).Return(7770.0, errors.New("gauge update error"))

				return mockStorage
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "counter update error",
			method: http.MethodPost,
			url:    "/update/counter/PollCount/30",
			setupMocks: func(ctrl *gomock.Controller) *MockHandlerStorage {
				mockStorage := NewMockHandlerStorage(ctrl)
				mockStorage.EXPECT().UpdateCounter("PollCount", int64(30)).Return(int64(30), errors.New("counter update error"))
				return mockStorage
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := tt.setupMocks(ctrl)

			h := &Handler{
				storage: storage,
			}

			r := chi.NewRouter()
			r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testutil.TestRequest(t, ts, tt.method, tt.url, "")
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected status code %d, got %d", tt.expectedStatusCode, resp.StatusCode)
		})
	}
}
