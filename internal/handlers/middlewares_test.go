package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoggingMiddleware(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
		expectedFormat string
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			path:           "/test/path",
			expectedStatus: http.StatusOK,
			expectedBody:   "base handler called",
			expectedFormat: "Incoming request: %s %s",
		},
		{
			name:           "POST request",
			method:         http.MethodPost,
			path:           "/update/counter/metric/1",
			expectedStatus: http.StatusOK,
			expectedBody:   "base handler called",
			expectedFormat: "Incoming request: %s %s",
		},
		{
			name:           "PUT request",
			method:         http.MethodPut,
			path:           "/api/v1/metrics",
			expectedStatus: http.StatusOK,
			expectedBody:   "base handler called",
			expectedFormat: "Incoming request: %s %s",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLog := NewMockhandlersLogger(ctrl)

			mockLog.EXPECT().Info(tc.expectedFormat, tc.method, tc.path).Times(1)

			baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("base handler called"))
			})

			handler := loggingMiddleware(mockLog)(baseHandler)

			req, err := http.NewRequest(tc.method, tc.path, nil)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, tc.expectedBody, recorder.Body.String())
		})
	}
}
