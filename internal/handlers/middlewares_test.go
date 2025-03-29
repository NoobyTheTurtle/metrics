package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddleware(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
		expectedLog    string
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			path:           "/test/path",
			expectedStatus: http.StatusOK,
			expectedBody:   "base handler called",
			expectedLog:    "[INFO] Incoming request: GET /test/path",
		},
		{
			name:           "POST request",
			method:         http.MethodPost,
			path:           "/update/counter/metric/1",
			expectedStatus: http.StatusOK,
			expectedBody:   "base handler called",
			expectedLog:    "[INFO] Incoming request: POST /update/counter/metric/1",
		},
		{
			name:           "PUT request",
			method:         http.MethodPut,
			path:           "/api/v1/metrics",
			expectedStatus: http.StatusOK,
			expectedBody:   "base handler called",
			expectedLog:    "[INFO] Incoming request: PUT /api/v1/metrics",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLog := logger.NewMockLogger()

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

			logOutput := mockLog.GetOutput()
			assert.True(t, strings.Contains(logOutput, tc.expectedLog),
				"Log output should contain \"%s\", got: %s", tc.expectedLog, logOutput)
		})
	}
}
