package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/NoobyTheTurtle/metrics/internal/hash"
)

func TestHashValidator(t *testing.T) {
	secretKey := "testSecret"
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	tests := []struct {
		name                 string
		key                  string
		requestBody          string
		hashHeader           string
		provideHeader        bool
		calculateDynamicHash bool
		expectedStatus       int
		expectNextCall       bool
	}{
		{
			name:                 "empty key - should pass",
			key:                  "",
			requestBody:          "test body",
			provideHeader:        true,
			hashHeader:           "any",
			calculateDynamicHash: false,
			expectedStatus:       http.StatusOK,
			expectNextCall:       true,
		},
		{
			name:                 "mismatched hash - should fail",
			key:                  secretKey,
			requestBody:          "test body",
			provideHeader:        true,
			hashHeader:           "wronghash",
			calculateDynamicHash: false,
			expectedStatus:       http.StatusBadRequest,
			expectNextCall:       false,
		},
		{
			name:                 "valid hash - should pass",
			key:                  secretKey,
			requestBody:          "test body",
			provideHeader:        true,
			calculateDynamicHash: true,
			expectedStatus:       http.StatusOK,
			expectNextCall:       true,
		},
		{
			name:                 "empty body with valid hash - should pass",
			key:                  secretKey,
			requestBody:          "",
			provideHeader:        true,
			calculateDynamicHash: true,
			expectedStatus:       http.StatusOK,
			expectNextCall:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			logger := NewMockMiddlewareLogger(ctrl)
			var nextCalled bool
			handlerToTest := HashValidator(tt.key, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				testHandler.ServeHTTP(w, r)
			}))

			reqBody := bytes.NewBufferString(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", reqBody)

			if tt.name == "no hash header - should fail" {
				logger.EXPECT().Info(
					"Request without HashSHA256 header from %s for %s",
					gomock.Any(),
					gomock.Any(),
				).Times(1)
			}
			if tt.name == "mismatched hash - should fail" {
				logger.EXPECT().Info(
					"Hash mismatch for request from %s for %s. Incoming: %s, Calculated: %s",
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Times(1)
			}

			if tt.provideHeader {
				headerValue := tt.hashHeader
				if tt.calculateDynamicHash {
					calculatedHash, err := hash.CalculateSHA256([]byte(tt.requestBody), tt.key)
					require.NoError(t, err)
					headerValue = calculatedHash
				}
				req.Header.Set("HashSHA256", headerValue)
			}

			rr := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectNextCall, nextCalled)

			if tt.expectNextCall && tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "OK", rr.Body.String())
			}
		})
	}
}

func TestHashAppender(t *testing.T) {
	defaultTestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	tests := []struct {
		name             string
		key              string
		handler          http.HandlerFunc
		expectedHash     bool
		expectedResponse string
		expectedStatus   int
	}{
		{
			name:             "empty key - no hash added",
			key:              "",
			handler:          defaultTestHandler,
			expectedHash:     false,
			expectedResponse: `{"status":"ok"}`,
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "with key - hash added",
			key:              "testSecretAppender",
			handler:          defaultTestHandler,
			expectedHash:     true,
			expectedResponse: `{"status":"ok"}`,
			expectedStatus:   http.StatusOK,
		},
		{
			name: "handler writes nothing",
			key:  "testSecretAppender",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
			expectedHash:     true,
			expectedResponse: "",
			expectedStatus:   http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			logger := NewMockMiddlewareLogger(ctrl)

			handlerToTest := HashAppender(tt.key, logger)(tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			responseBody, err := io.ReadAll(rr.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResponse, string(responseBody))

			if tt.expectedHash {
				actualHash := rr.Header().Get("HashSHA256")
				require.NotEmpty(t, actualHash, "HashSHA256 header should be present")

				expectedCalculatedHash, err := hash.CalculateSHA256(responseBody, tt.key)
				require.NoError(t, err)
				assert.Equal(t, expectedCalculatedHash, actualHash)
			} else {
				assert.Empty(t, rr.Header().Get("HashSHA256"), "HashSHA256 header should not be present")
			}
		})
	}
}

func TestHashValidator_MissingKey_PassThrough(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := NewMockMiddlewareLogger(ctrl)

	var nextCalled bool
	handler := HashValidator("", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("test body"))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, nextCalled)
}

func TestHashValidator_InvalidHash_Error(t *testing.T) {
	secretKey := "testSecret"
	ctrl := gomock.NewController(t)
	logger := NewMockMiddlewareLogger(ctrl)

	logger.EXPECT().Info(
		"Hash mismatch for request from %s for %s. Incoming: %s, Calculated: %s",
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Times(1)

	var nextCalled bool
	handler := HashValidator(secretKey, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("test body"))
	req.Header.Set("HashSHA256", "invalid-hash")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.False(t, nextCalled)
}

func TestHashValidator_MissingHeader_PassThrough(t *testing.T) {
	secretKey := "testSecret"
	ctrl := gomock.NewController(t)
	logger := NewMockMiddlewareLogger(ctrl)

	var nextCalled bool
	handler := HashValidator(secretKey, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("test body"))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, nextCalled)
}

func TestHashAppender_NoKey_PassThrough(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := NewMockMiddlewareLogger(ctrl)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("response body"))
	})

	handler := HashAppender("", logger)(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "response body", rr.Body.String())
	assert.Empty(t, rr.Header().Get("HashSHA256"))
}

func TestHashAppender_ErrorResponse_NoHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := NewMockMiddlewareLogger(ctrl)

	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})

	handler := HashAppender("secret-key", logger)(errorHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("HashSHA256"))
}

func TestHashWriter_Methods(t *testing.T) {
	originalWriter := httptest.NewRecorder()
	hw := &hashWriter{
		originalWriter: originalWriter,
		body:           bytes.NewBuffer([]byte{}),
		statusCode:     0,
	}

	t.Run("Write", func(t *testing.T) {
		data := []byte("test data")
		n, err := hw.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, "test data", hw.body.String())
	})

	t.Run("Header", func(t *testing.T) {
		header := hw.Header()
		assert.Equal(t, originalWriter.Header(), header)
	})

	t.Run("WriteHeader", func(t *testing.T) {
		hw.WriteHeader(http.StatusCreated)
		assert.Equal(t, http.StatusCreated, hw.statusCode)
	})
}
