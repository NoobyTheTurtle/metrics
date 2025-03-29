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
