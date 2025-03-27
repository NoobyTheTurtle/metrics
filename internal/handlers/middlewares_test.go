package handlers

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConveyor(t *testing.T) {
	tests := []struct {
		name             string
		middlewares      []Middleware
		expectedResponse string
	}{
		{
			name:             "without middlewares",
			middlewares:      []Middleware{},
			expectedResponse: "base handler called",
		},
		{
			name: "with one middleware",
			middlewares: []Middleware{
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("middleware1 before;"))
						next.ServeHTTP(w, r)
						w.Write([]byte(";middleware1 after"))
					})
				},
			},
			expectedResponse: "middleware1 before;base handler called;middleware1 after",
		},
		{
			name: "with multiple middlewares",
			middlewares: []Middleware{
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("middleware1 before;"))
						next.ServeHTTP(w, r)
						w.Write([]byte(";middleware1 after"))
					})
				},
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("middleware2 before;"))
						next.ServeHTTP(w, r)
						w.Write([]byte(";middleware2 after"))
					})
				},
			},
			expectedResponse: "middleware2 before;middleware1 before;base handler called;middleware1 after;middleware2 after",
		},
	}

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("base handler called"))
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			handler := conveyor(baseHandler, tt.middlewares...)

			handler.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.Equal(t, tt.expectedResponse, recorder.Body.String())
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(os.Stderr)

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("base handler called"))
	})

	handler := loggingMiddleware(baseHandler)

	req, err := http.NewRequest(http.MethodGet, "/test/path", nil)
	require.NoError(t, err)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "base handler called", recorder.Body.String())

	logOutput := logBuffer.String()
	assert.True(t, strings.Contains(logOutput, "Incoming request: GET /test/path"), "Log output should contain request info, got: %s", logOutput)
}
