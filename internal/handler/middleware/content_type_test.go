package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeMiddleware(t *testing.T) {
	testCases := []struct {
		name           string
		contentType    string
		method         string
		path           string
		responseStatus int
		responseBody   string
		handlerFunc    func(w http.ResponseWriter, r *http.Request)
	}{
		{
			name:           "JSON content type",
			contentType:    "application/json",
			method:         http.MethodGet,
			path:           "/test",
			responseStatus: http.StatusOK,
			responseBody:   `{"status":"ok"}`,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			},
		},
		{
			name:           "Text content type",
			contentType:    "text/plain",
			method:         http.MethodGet,
			path:           "/test",
			responseStatus: http.StatusOK,
			responseBody:   "test1 2\ntest2 345",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test1 2\ntest2 345"))
			},
		},
		{
			name:           "HTML content type",
			contentType:    "text/html",
			method:         http.MethodGet,
			path:           "/",
			responseStatus: http.StatusOK,
			responseBody:   "<html><body>Hello</body></html>",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("<html><body>Hello</body></html>"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseHandler := http.HandlerFunc(tc.handlerFunc)
			handler := ContentTypeMiddleware(tc.contentType)(baseHandler)

			req, err := http.NewRequest(tc.method, tc.path, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.responseStatus, recorder.Code)

			assert.Equal(t, tc.responseBody, recorder.Body.String())

			assert.Equal(t, tc.contentType, recorder.Header().Get("Content-Type"))
		})
	}
}
