package handlers

import (
	"bytes"
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
		requestBody    []byte
		responseStatus int
		responseBody   string
		handlerFunc    func(w http.ResponseWriter, r *http.Request)
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			path:           "/test/path",
			responseStatus: http.StatusOK,
			responseBody:   "get handler response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("get handler response"))
			},
		},
		{
			name:           "POST request with body",
			method:         http.MethodPost,
			path:           "/api/v1/update",
			requestBody:    []byte(`{"metric":"test","value":123}`),
			responseStatus: http.StatusCreated,
			responseBody:   "post handler response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("post handler response"))
			},
		},
		{
			name:           "Error response",
			method:         http.MethodPut,
			path:           "/api/v1/error",
			responseStatus: http.StatusBadRequest,
			responseBody:   "error response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("error response"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLog := NewMockHandlersLogger(ctrl)

			mockLog.EXPECT().Info(
				"uri=%s method=%s status=%d duration=%s size=%d",
				gomock.Eq(""),
				tc.method,
				tc.responseStatus,
				gomock.Any(),
				len(tc.responseBody),
			).Times(1)

			baseHandler := http.HandlerFunc(tc.handlerFunc)
			handler := loggingMiddleware(mockLog)(baseHandler)

			req, err := http.NewRequest(tc.method, tc.path, bytes.NewBuffer(tc.requestBody))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.responseStatus, recorder.Code)
			assert.Equal(t, tc.responseBody, recorder.Body.String())
		})
	}
}
