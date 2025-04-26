package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/NoobyTheTurtle/metrics/internal/handler/html"
	"github.com/NoobyTheTurtle/metrics/internal/handler/json"
	"github.com/NoobyTheTurtle/metrics/internal/handler/plain"
	"github.com/NoobyTheTurtle/metrics/internal/testutil"
)

func TestNewRouter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockMetricStorage(ctrl)
	mockLogger := NewMockRouterLogger(ctrl)
	mockDBPinger := NewMockDBPinger(ctrl)

	router := NewRouter(mockStorage, mockLogger, mockDBPinger)

	assert.NotNil(t, router)
	assert.NotNil(t, router.router)
	assert.Equal(t, mockStorage, router.storage)
	assert.Equal(t, mockLogger, router.logger)
	assert.NotNil(t, router.htmlHandler)
	assert.NotNil(t, router.plainHandler)
	assert.NotNil(t, router.jsonHandler)
	assert.NotNil(t, router.pingHandler)
}

func TestRouter_Handler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockMetricStorage(ctrl)
	mockLogger := NewMockRouterLogger(ctrl)
	mockDBPinger := NewMockDBPinger(ctrl)

	router := NewRouter(mockStorage, mockLogger, mockDBPinger)
	handler := router.Handler()

	assert.NotNil(t, handler)
	assert.IsType(t, chi.NewRouter(), handler)
}

func TestRouter_Routes(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		path               string
		requestBody        string
		contentType        string
		setupMocks         func(*gomock.Controller) (*MockMetricStorage, *MockRouterLogger, *MockDBPinger)
		expectedStatusCode int
	}{
		{
			name:        "Ping route successful",
			method:      http.MethodGet,
			path:        "/ping",
			contentType: "text/plain",
			setupMocks: func(ctrl *gomock.Controller) (*MockMetricStorage, *MockRouterLogger, *MockDBPinger) {
				mockStorage := NewMockMetricStorage(ctrl)
				mockLogger := NewMockRouterLogger(ctrl)
				mockDBPinger := NewMockDBPinger(ctrl)

				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
				mockDBPinger.EXPECT().Ping(gomock.Any()).Return(nil)

				return mockStorage, mockLogger, mockDBPinger
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "Ping route database error",
			method:      http.MethodGet,
			path:        "/ping",
			contentType: "text/plain",
			setupMocks: func(ctrl *gomock.Controller) (*MockMetricStorage, *MockRouterLogger, *MockDBPinger) {
				mockStorage := NewMockMetricStorage(ctrl)
				mockLogger := NewMockRouterLogger(ctrl)
				mockDBPinger := NewMockDBPinger(ctrl)

				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
				mockDBPinger.EXPECT().Ping(gomock.Any()).Return(errors.New("database error"))
				// mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)

				return mockStorage, mockLogger, mockDBPinger
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "HTML route",
			method:      http.MethodGet,
			path:        "/",
			contentType: html.ContentTypeValue,
			setupMocks: func(ctrl *gomock.Controller) (*MockMetricStorage, *MockRouterLogger, *MockDBPinger) {
				mockStorage := NewMockMetricStorage(ctrl)
				mockLogger := NewMockRouterLogger(ctrl)
				mockDBPinger := NewMockDBPinger(ctrl)

				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
				mockStorage.EXPECT().GetAllGauges().Return(map[string]float64{})
				mockStorage.EXPECT().GetAllCounters().Return(map[string]int64{})

				return mockStorage, mockLogger, mockDBPinger
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "Plain value route",
			method:      http.MethodGet,
			path:        "/value/gauge/test",
			contentType: plain.ContentTypeValue,
			setupMocks: func(ctrl *gomock.Controller) (*MockMetricStorage, *MockRouterLogger, *MockDBPinger) {
				mockStorage := NewMockMetricStorage(ctrl)
				mockLogger := NewMockRouterLogger(ctrl)
				mockDBPinger := NewMockDBPinger(ctrl)

				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
				mockStorage.EXPECT().GetGauge("test").Return(float64(10.5), true)

				return mockStorage, mockLogger, mockDBPinger
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "Plain update route",
			method:      http.MethodPost,
			path:        "/update/gauge/test/15.5",
			contentType: plain.ContentTypeValue,
			setupMocks: func(ctrl *gomock.Controller) (*MockMetricStorage, *MockRouterLogger, *MockDBPinger) {
				mockStorage := NewMockMetricStorage(ctrl)
				mockLogger := NewMockRouterLogger(ctrl)
				mockDBPinger := NewMockDBPinger(ctrl)

				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
				mockStorage.EXPECT().UpdateGauge("test", float64(15.5)).Return(float64(15.5), nil)

				return mockStorage, mockLogger, mockDBPinger
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "JSON update route",
			method:      http.MethodPost,
			path:        "/update/",
			requestBody: `{"id":"test","type":"gauge","value":12.3}`,
			contentType: json.ContentTypeValue,
			setupMocks: func(ctrl *gomock.Controller) (*MockMetricStorage, *MockRouterLogger, *MockDBPinger) {
				mockStorage := NewMockMetricStorage(ctrl)
				mockLogger := NewMockRouterLogger(ctrl)
				mockDBPinger := NewMockDBPinger(ctrl)

				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
				mockStorage.EXPECT().UpdateGauge("test", 12.3).Return(12.3, nil)

				return mockStorage, mockLogger, mockDBPinger
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "JSON value route",
			method:      http.MethodPost,
			path:        "/value/",
			requestBody: `{"id":"test","type":"counter"}`,
			contentType: json.ContentTypeValue,
			setupMocks: func(ctrl *gomock.Controller) (*MockMetricStorage, *MockRouterLogger, *MockDBPinger) {
				mockStorage := NewMockMetricStorage(ctrl)
				mockLogger := NewMockRouterLogger(ctrl)
				mockDBPinger := NewMockDBPinger(ctrl)

				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
				mockStorage.EXPECT().GetCounter("test").Return(int64(42), true)

				return mockStorage, mockLogger, mockDBPinger
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "Route not found",
			method:      http.MethodGet,
			path:        "/not-found",
			contentType: "text/plain",
			setupMocks: func(ctrl *gomock.Controller) (*MockMetricStorage, *MockRouterLogger, *MockDBPinger) {
				mockStorage := NewMockMetricStorage(ctrl)
				mockLogger := NewMockRouterLogger(ctrl)
				mockDBPinger := NewMockDBPinger(ctrl)

				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

				return mockStorage, mockLogger, mockDBPinger
			},
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage, mockLogger, mockDBPinger := tt.setupMocks(ctrl)
			router := NewRouter(mockStorage, mockLogger, mockDBPinger)

			ts := httptest.NewServer(router.Handler())
			defer ts.Close()

			req, err := http.NewRequest(tt.method, ts.URL+tt.path, nil)
			req.Header.Set("Content-Type", tt.contentType)
			require.NoError(t, err)

			resp, _ := testutil.TestRequest(t, ts, tt.method, tt.path, tt.requestBody)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected status code %d, got %d for %s %s",
				tt.expectedStatusCode, resp.StatusCode, tt.method, tt.path)
		})
	}
}
